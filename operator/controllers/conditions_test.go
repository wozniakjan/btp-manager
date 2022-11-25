package controllers

import (
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testing"
)

func TestNewConditionForReason(t *testing.T) {
	t.Run("should create new condition for given predefined Reason with status True", func(t *testing.T) {
		condition := NewConditionForReason("ReconcileSucceeded", "Ready to process")
		assert.Equal(t, "Ready", condition.Type)
		assert.Equal(t, metav1.ConditionTrue, condition.Status)
		assert.Equal(t, "Ready to process", condition.Message)
		assert.Equal(t, "ReconcileSucceeded", condition.Reason)
	})
	t.Run("should create new condition for given predefined Reason with status False", func(t *testing.T) {
		condition := NewConditionForReason("OlderCRExists", "Other CR is elected a leader")
		assert.Equal(t, "Ready", condition.Type)
		assert.Equal(t, metav1.ConditionFalse, condition.Status)
		assert.Equal(t, "Other CR is elected a leader", condition.Message)
		assert.Equal(t, "OlderCRExists", condition.Reason)
	})
	t.Run("should not create new condition for not predefined Reason", func(t *testing.T) {
		condition := NewConditionForReason("non-existing-reason", "Ready to process")
		assert.Nil(t, condition)
	})
}

func TestSetStatusCondition(t *testing.T) {
	t.Run("should add single condition to the empty set", func(t *testing.T) {
		condition := NewConditionForReason("ReconcileSucceeded", "Ready to process")

		btpOperator := createBtpOperator()
		SetStatusCondition(&btpOperator.Status.Conditions, *condition)

		assert.Equal(t, 1, len(btpOperator.Status.Conditions))
		assert.Equal(t, "Ready", btpOperator.Status.Conditions[0].Type)
		assert.Equal(t, metav1.ConditionTrue, btpOperator.Status.Conditions[0].Status)
		assert.Equal(t, "Ready to process", btpOperator.Status.Conditions[0].Message)
		assert.Equal(t, "ReconcileSucceeded", btpOperator.Status.Conditions[0].Reason)
	})
	t.Run("should add the condition with the same type only once", func(t *testing.T) {
		condition := NewConditionForReason("ReconcileSucceeded", "Ready to process")

		btpOperator := createBtpOperator()
		SetStatusCondition(&btpOperator.Status.Conditions, *condition)
		SetStatusCondition(&btpOperator.Status.Conditions, *condition)

		assert.Equal(t, 1, len(btpOperator.Status.Conditions))
		assert.Equal(t, "Ready", btpOperator.Status.Conditions[0].Type)
		assert.Equal(t, metav1.ConditionTrue, btpOperator.Status.Conditions[0].Status)
		assert.Equal(t, "Ready to process", btpOperator.Status.Conditions[0].Message)
		assert.Equal(t, "ReconcileSucceeded", btpOperator.Status.Conditions[0].Reason)
	})
	t.Run("should update conditions of the same type with new values", func(t *testing.T) {
		precondition := NewConditionForReason("ReconcileSucceeded", "Ready to process")
		postcondition := NewConditionForReason("MissingSecret", "No secret found")
		btpOperator := createBtpOperator()
		SetStatusCondition(&btpOperator.Status.Conditions, *precondition)
		SetStatusCondition(&btpOperator.Status.Conditions, *postcondition)

		assert.Equal(t, 1, len(btpOperator.Status.Conditions))
		assert.Equal(t, "Ready", btpOperator.Status.Conditions[0].Type)
		assert.Equal(t, metav1.ConditionFalse, btpOperator.Status.Conditions[0].Status)
		assert.Equal(t, "No secret found", btpOperator.Status.Conditions[0].Message)
		assert.Equal(t, "MissingSecret", btpOperator.Status.Conditions[0].Reason)
	})
}
