package controllers

import (
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Reason string

const (
	ReconcileSucceeded                 Reason = "ReconcileSucceeded"
	ReconcileFailed                    Reason = "ReconcileFailed"
	Initialized                        Reason = "Initialized"
	Processing                         Reason = "Processing"
	OlderCRExists                      Reason = "OlderCRExists"
	ChartInstallFailed                 Reason = "ChartInstallFailed"
	ConsistencyCheckFailed             Reason = "ConsistencyCheckFailed"
	MissingSecret                      Reason = "MissingSecret"
	InvalidSecret                      Reason = "InvalidSecret"
	HardDeleting                       Reason = "HardDeleting"
	ResourceRemovalFailed              Reason = "ResourceRemovalFailed"
	SoftDeleting                       Reason = "SoftDeleting"
	Updated                            Reason = "Updated"
	UpdateCheck                        Reason = "UpdateCheck"
	UpdateCheckSucceeded               Reason = "UpdateCheckSucceeded"
	InconsistentChart                  Reason = "InconsistentChart"
	UpdateDone                         Reason = "UpdateDone"
	PreparingInstallInfoFailed         Reason = "PreparingInstallInfoFailed"
	ChartPathEmpty                     Reason = "ChartPathEmpty"
	DeletionOfOrphanedResourcesFailed  Reason = "DeletionOfOrphanedResourcesFailed"
	StoringChartDetailsFailed          Reason = "StoringChartDetailsFailed"
	GettingConfigMapFailed             Reason = "GettingConfigMapFailed"
	CreatingObjectsFromManifestsFailed Reason = "CreatingObjectsFromManifestsFailed"
	PreparingModuleResourcesFailed     Reason = "PreparingModuleResourcesFailed"
	ProvisioningFailed                 Reason = "ProvisioningFailed"
	UpdateFailed                       Reason = "UpdateFailed"
	ReadyType                                 = "Ready"
)

type TypeAndStatus struct {
	Status metav1.ConditionStatus
	Type   string
}

var Ready = TypeAndStatus{
	Status: metav1.ConditionTrue,
	Type:   ReadyType,
}

var NotReady = TypeAndStatus{
	Status: metav1.ConditionFalse,
	Type:   ReadyType,
}

var Reasons = map[Reason]TypeAndStatus{
	ReconcileSucceeded:                 Ready,
	UpdateDone:                         Ready,
	UpdateCheckSucceeded:               Ready,
	ReconcileFailed:                    NotReady,
	Updated:                            NotReady,
	Initialized:                        NotReady,
	ChartInstallFailed:                 NotReady,
	ConsistencyCheckFailed:             NotReady,
	Processing:                         NotReady,
	OlderCRExists:                      NotReady,
	MissingSecret:                      NotReady,
	InvalidSecret:                      NotReady,
	HardDeleting:                       NotReady,
	ResourceRemovalFailed:              NotReady,
	SoftDeleting:                       NotReady,
	UpdateCheck:                        NotReady,
	InconsistentChart:                  NotReady,
	PreparingInstallInfoFailed:         NotReady,
	ChartPathEmpty:                     NotReady,
	DeletionOfOrphanedResourcesFailed:  NotReady,
	StoringChartDetailsFailed:          NotReady,
	GettingConfigMapFailed:             NotReady,
	CreatingObjectsFromManifestsFailed: NotReady,
	PreparingModuleResourcesFailed:     NotReady,
	ProvisioningFailed:                 NotReady,
	UpdateFailed:                       NotReady,
}

func ConditionFromExistingReason(reason Reason, message string) *metav1.Condition {
	typeAndStatus, found := Reasons[reason]
	if found {
		return &metav1.Condition{
			Status:             typeAndStatus.Status,
			Reason:             string(reason),
			Message:            message,
			Type:               typeAndStatus.Type,
			ObservedGeneration: 0,
		}
	}
	return nil
}

// This is required because of difference between Conditions declarations
// In BtpOperator we have Status.Conditions []*Condition instead of Status.Conditions []Condition
func SetStatusCondition(conditions *[]*metav1.Condition, newCondition metav1.Condition) {
	conditionsCnt := len(*conditions)
	var conditionsArray = make([]metav1.Condition, conditionsCnt, conditionsCnt)
	for i := 0; i < conditionsCnt; i++ {
		conditionsArray[i] = *(*conditions)[i]
	}

	apimeta.SetStatusCondition(&conditionsArray, newCondition)

	for i := 0; i < conditionsCnt; i++ {
		(*conditions)[i] = &conditionsArray[i]
	}
	if len(conditionsArray) > conditionsCnt {
		*conditions = append(*conditions, &metav1.Condition{})
		(*conditions)[conditionsCnt] = &conditionsArray[conditionsCnt]
	}
}
