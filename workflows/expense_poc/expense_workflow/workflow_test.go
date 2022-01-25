package expense_workflow

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_WorkflowWithMockActivities() {
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(CreateExpenseActivity, mock.Anything, mock.Anything).Return(nil).Once()
	env.OnActivity(WaitForDecisionActivity, mock.Anything, mock.Anything).Return("APPROVED", nil).Once()
	env.OnActivity(PaymentActivity, mock.Anything, mock.Anything).Return(nil).Once()

	env.ExecuteWorkflow(SampleExpenseWorkflow, "test-expense-id")

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var workflowResult string
	err := env.GetWorkflowResult(&workflowResult)
	s.NoError(err)
	s.Equal("COMPLETED", workflowResult)
	env.AssertExpectations(s.T())
}
