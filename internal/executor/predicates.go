package executor

import (
	"github.com/vyper-tooling/pi/internal/conditions"
)

// RuntimeEnv is an alias for conditions.RuntimeEnv. It captures the runtime
// values needed for predicate resolution. Kept here for backward compatibility
// with executor consumers (e.g. doctor, coverage_gaps_test).
type RuntimeEnv = conditions.RuntimeEnv

// DefaultRuntimeEnv returns a RuntimeEnv backed by the real OS.
// Delegates to conditions.DefaultRuntimeEnv().
func DefaultRuntimeEnv() *RuntimeEnv {
	return conditions.DefaultRuntimeEnv()
}

// ResolvePredicates resolves a list of predicate names to boolean values.
// Delegates to conditions.ResolvePredicates().
func ResolvePredicates(predicateNames []string, repoRoot string) (map[string]bool, error) {
	return conditions.ResolvePredicates(predicateNames, repoRoot)
}

// ResolvePredicatesWithEnv is the testable variant that accepts an injected RuntimeEnv.
// Delegates to conditions.ResolvePredicatesWithEnv().
func ResolvePredicatesWithEnv(predicateNames []string, repoRoot string, env *RuntimeEnv) (map[string]bool, error) {
	return conditions.ResolvePredicatesWithEnv(predicateNames, repoRoot, env)
}

// ValidatePredicateName checks whether a predicate name is statically valid.
// Delegates to conditions.ValidatePredicateName().
func ValidatePredicateName(name string) error {
	return conditions.ValidatePredicateName(name)
}

// ValidateConditionExpr statically validates an if: expression.
// Delegates to conditions.ValidateConditionExpr().
func ValidateConditionExpr(expr string) error {
	return conditions.ValidateConditionExpr(expr)
}
