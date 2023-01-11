package controllers

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kyma-project/btp-manager/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func TestProvision(t *testing.T) {
	ctx, _ := context.WithCancel(context.TODO())
	log.Default().SetFlags(log.Lmicroseconds)

	log.Println("bootstrapping test environment")
	testEnv := &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	if err != nil {
		log.Fatal("failed to create test env", err)
	}

	log.Println("setup manager")
	err = v1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatal("failed to init scheme", err)
	}

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		log.Fatal("failed to create k8s client", err)
	}

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		log.Fatal("failed to create manager", err)
	}

	reconciler := &BtpOperatorReconciler{
		Client:                k8sManager.GetClient(),
		Scheme:                k8sManager.GetScheme(),
		WaitForChartReadiness: false,
	}
	//k8sClientFromManager := k8sManager.GetClient()
	HardDeleteTimeout = time.Second
	HardDeleteCheckInterval = time.Millisecond * 100
	ChartPath = "/tmp/module-chart"
	if err := k8sClient.Create(ctx, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kyma-system"}}); err != nil {
		log.Fatal("failed to create secret", err)
	}

	err = reconciler.SetupWithManager(k8sManager)
	if err != nil {
		log.Fatal("failed to init manager", err)
	}

	log.Println("setting informers")
	informer, err := k8sManager.GetCache().GetInformer(ctx, &v1alpha1.BtpOperator{})
	if err != nil {
		log.Fatal("failed to setup informer", err)
	}

	updates := make(chan *v1alpha1.BtpOperator, 100)
	handler := func(obj any, x ...string) {
		cr, ok := obj.(*v1alpha1.BtpOperator)
		if ok {
			for _, l := range x {
				if cr.Labels == nil {
					cr.Labels = make(map[string]string)
				}
				cr.Labels["state"] = l
			}
			updates <- cr
		}
	}
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(o any) { handler(o, "add") },
		UpdateFunc: func(o, n any) { handler(n, "update") },
		DeleteFunc: func(o any) { handler(o, "delete") },
	})

	go func() {
		if err := k8sManager.Start(ctx); err != nil {
			log.Fatal("failed to init manager", err)
		}
	}()

	log.Println("wait for cache sync")
	k8sManager.GetCache().WaitForCacheSync(ctx)

	log.Println("create btp operator")
	secret, err := createCorrectSecretFromYaml()
	if err != nil {
		log.Fatal("failed to init secret", err)
	}
	if err := k8sClient.Create(ctx, secret); err != nil {
		log.Fatal("failed to create secret", err)
	}
	cr := createBtpOperator()
	if err := k8sClient.Create(ctx, cr); err != nil {
		log.Fatal("failed to create btp operator resource", err)
	}

	log.Println("waiting for btp operator to become ready")
	listen := true
	for listen {
		cr := <-updates
		log.Println("conditions", len(cr.Status.Conditions))
		for _, c := range cr.Status.Conditions {
			log.Println(c)
			if c.Type == "Ready" && c.Status == metav1.ConditionTrue {
				listen = false
			}
		}
	}

	log.Println("creating resources")
	createResource(k8sClient, schema.GroupVersionKind{Group: "services.cloud.sap.com", Version: "v1", Kind: "ServiceInstance"}, "default", "instance")
	ensureResourceExists(k8sClient, instanceGvk)

	createResource(k8sClient, schema.GroupVersionKind{Group: "services.cloud.sap.com", Version: "v1", Kind: "ServiceBinding"}, "default", "binding")
	ensureResourceExists(k8sClient, bindingGvk)

	log.Println("delete cr")
	err = k8sClient.Get(ctx, types.NamespacedName{Namespace: cr.Namespace, Name: cr.Name}, cr)
	if err != nil {
		log.Fatalf("expected CR to be exist: %v", err)
	}
	if err := k8sClient.Delete(ctx, cr); err != nil {
		log.Fatalf("failed to delete cr: %v", err)
	}
	listen = true
	for listen {
		cr := <-updates
		log.Println("when delete conditions", len(cr.Status.Conditions))
		for _, c := range cr.Status.Conditions {
			log.Println(cr.Labels, cr.DeletionTimestamp, cr.Finalizers, c)
		}
		if cr.Labels["state"] == "delete" {
			listen = false
		}
	}
	err = k8sClient.Get(ctx, types.NamespacedName{Namespace: cr.Namespace, Name: cr.Name}, cr)
	if err == nil || !errors.IsNotFound(err) {
		log.Fatalf("expected CR to be deleted: %v", err)
	}

	log.Println("done")
}

func ensureResourceExists(k8sClient client.Client, gvk schema.GroupVersionKind) {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(gvk)
	err := k8sClient.List(context.TODO(), list)
	if err != nil {
		log.Fatalf("failed to list resource %v: %v", gvk, err)
	}
	if len(list.Items) != 1 {
		log.Fatalf("expected 1 %v found %v", gvk, len(list.Items))
	}
}

func createResource(k8sClient client.Client, gvk schema.GroupVersionKind, namespace string, name string) *unstructured.Unstructured {
	object := &unstructured.Unstructured{}
	object.SetGroupVersionKind(gvk)
	object.SetNamespace(namespace)
	object.SetName(name)
	object.SetFinalizers([]string{"xyz"})
	kind := object.GetObjectKind().GroupVersionKind().Kind
	if kind == "ServiceInstance" {
		unstructured.SetNestedField(object.Object, "test-service", "spec", "serviceOfferingName")
		unstructured.SetNestedField(object.Object, "test-plan", "spec", "servicePlanName")
		unstructured.SetNestedField(object.Object, "test-service-instance-external", "spec", "externalName")
	}
	if kind == "ServiceBinding" {
		unstructured.SetNestedField(object.Object, "test-service-instance", "spec", "serviceInstanceName")
		unstructured.SetNestedField(object.Object, "test-binding-external", "spec", "externalName")
		unstructured.SetNestedField(object.Object, "test-service-binding-secret", "spec", "secretName")
	}
	err := k8sClient.Create(context.TODO(), object)
	if err != nil {
		log.Fatalf("failed to create resource %v", err)
	}

	return object
}

func createCorrectSecretFromYaml() (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	data, err := os.ReadFile("testdata/test-secret.yaml")
	if err != nil {
		return nil, fmt.Errorf("while reading the required Secret YAML: %w", err)
	}
	err = yaml.Unmarshal(data, secret)
	if err != nil {
		return nil, fmt.Errorf("while unmarshalling Secret YAML to struct: %w", err)
	}

	return secret, nil
}

func createBtpOperator() *v1alpha1.BtpOperator {
	return &v1alpha1.BtpOperator{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "btp-operator-test",
			Namespace: "kyma-system",
		},
	}
}
