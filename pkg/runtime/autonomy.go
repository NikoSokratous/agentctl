package runtime

// RequiresApproval returns true if the given context requires human approval
// at the configured autonomy level.
func RequiresApproval(level AutonomyLevel, policyRequiresApproval bool, riskScore float64) bool {
	switch level {
	case AutonomyManual:
		return true
	case AutonomyCautious:
		return riskScore >= 0.5 || policyRequiresApproval
	case AutonomyStandard:
		return policyRequiresApproval
	case AutonomyAutonomous, AutonomyUnrestricted:
		return false
	default:
		return true // safe default
	}
}

// EnforcesPolicy returns true if policy checks should run at this autonomy level.
func EnforcesPolicy(level AutonomyLevel) bool {
	return level != AutonomyUnrestricted
}
