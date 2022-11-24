package controllers

import (
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ReconcileSucceeded     = "ReconcileSucceeded"
	Initialized            = "Initialized"
	Processing             = "Processing"
	OlderCRExists          = "OlderCRExists"
	ChartInstallFailed     = "ChartInstallFailed"
	ConsistencyCheckFailed = "ConsistencyCheckFailed"
	MissingSecret          = "MissingSecret"
	InvalidSecret          = "InvalidSecret"
	HardDeleting           = "HardDeleting"
	DeprovisioningFailed   = "DeprovisioningFailed"
	ResourceRemovalFailed  = "ResourceRemovalFailed"
	HardDeleteFailed       = "HardDeleteFailed"
	SoftDeleteFailed       = "SoftDeleteFailed"
	SoftDeleting           = "SoftDeleting"
	Recovered              = "Recovered"
)

type TypeAndStatus struct {
	Status metav1.ConditionStatus
	Type   string
}

var Ready = TypeAndStatus{
	Status: metav1.ConditionTrue,
	Type:   "Ready",
}

var NotReady = TypeAndStatus{
	Status: metav1.ConditionFalse,
	Type:   "Ready",
}

var Reasons = map[string]TypeAndStatus{
	ReconcileSucceeded:     Ready,
	Initialized:            Ready,
	Recovered:              Ready,
	ChartInstallFailed:     NotReady,
	ConsistencyCheckFailed: NotReady,
	Processing:             NotReady,
	OlderCRExists:          NotReady,
	MissingSecret:          NotReady,
	InvalidSecret:          NotReady,
	HardDeleting:           NotReady,
	DeprovisioningFailed:   NotReady,
	ResourceRemovalFailed:  NotReady,
	HardDeleteFailed:       NotReady,
	SoftDeleteFailed:       NotReady,
	SoftDeleting:           NotReady,
}

func NewConditionByReason(reason string, message string) *metav1.Condition {
	typeAndStatus, found := Reasons[reason]
	if found {
		return &metav1.Condition{
			Status:             typeAndStatus.Status,
			Reason:             reason,
			Message:            message,
			Type:               typeAndStatus.Type,
			ObservedGeneration: 0,
		}
	}
	return nil
}

func SetStatusCondition(conditions []*metav1.Condition, newCondition metav1.Condition) {
	conditionsCnt := len(conditions)
	var conditionsArray = make([]metav1.Condition, conditionsCnt, conditionsCnt+1)
	for i := 0; i < conditionsCnt; i++ {
		conditionsArray[i] = *(conditions)[i]
	}
	apimeta.SetStatusCondition(&conditionsArray, newCondition)
	for i := 0; i < conditionsCnt; i++ {
		*(conditions)[i] = conditionsArray[i]
	}
	if len(conditionsArray) > conditionsCnt {
		conditions = append(conditions, &metav1.Condition{})
		*(conditions)[conditionsCnt] = conditionsArray[conditionsCnt]
	}
}
