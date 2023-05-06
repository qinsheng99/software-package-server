package dp

const (
	packageOperationLogActionUpdate  = "update"
	packageOperationLogActionReject  = "reject"
	packageOperationLogActionApprove = "approve"
	packageOperationLogActionAbandon = "abandon"
	packageOperationLogActionRerunci = "rerunci"
)

var (
	PackageOperationLogActionUpdate  = packageOperationLogAction(packageOperationLogActionUpdate)
	PackageOperationLogActionReject  = packageOperationLogAction(packageOperationLogActionReject)
	PackageOperationLogActionApprove = packageOperationLogAction(packageOperationLogActionApprove)
	PackageOperationLogActionAbandon = packageOperationLogAction(packageOperationLogActionAbandon)
	PackageOperationLogActionResunci = packageOperationLogAction(packageOperationLogActionRerunci)
)

type PackageOperationLogAction interface {
	PackageOperationLogAction() string
}

func NewPackageOperationLogAction(action string) PackageOperationLogAction {
	return packageOperationLogAction(action)
}

type packageOperationLogAction string

func (p packageOperationLogAction) PackageOperationLogAction() string {
	return string(p)
}
