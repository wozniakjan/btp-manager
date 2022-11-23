package controllers

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ReconcileSucceeded = "ReconcileSucceeded"
	Processing         = "Processing"
	OlderCRExists      = "OlderCRExists"
	ChartInstallFailed = "ChartInstallFailed"
	MissingSecret      = "MissingSecret"
	InvalidSecret      = "InvalidSecret"
	HardDeleting       = "HardDeleting"
	HardDeleteFailed   = "HardDeleteFailed"
	SoftDeleteFailed   = "SoftDeleteFailed"
	SoftDeleting       = "SoftDeleting"
	Recovered          = "Recovered"
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
	ReconcileSucceeded: Ready,
	Recovered:          Ready,
	ChartInstallFailed: NotReady,
	Processing:         NotReady,
	OlderCRExists:      NotReady,
	MissingSecret:      NotReady,
	InvalidSecret:      NotReady,
	HardDeleting:       NotReady,
	HardDeleteFailed:   NotReady,
	SoftDeleteFailed:   NotReady,
	SoftDeleting:       NotReady,
}

func NewConditionByReason(reason string, message string) *metav1.Condition {
	typeAndStatus, found := Reasons[reason]
	if found {
		return &metav1.Condition{
			Status:             typeAndStatus.Status,
			Reason:             reason,
			Message:            message,
			Type:               typeAndStatus.Type,
			ObservedGeneration: 0, //TODO handle generations
		}
	}
	return nil
}

func SetStatusCondition(conditions *[]*metav1.Condition, newCondition metav1.Condition) {
	//TODO add logic, to be discussed with MS and LJ
}
