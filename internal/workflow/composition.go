package workflow

import (
	"context"
	"slices"
	"sort"
	"strings"
	"time"

	"sdd.dev/sdd/internal/protocol/canonicaljson"
)

type Rigor string

const (
	RigorQuick    Rigor = "quick"
	RigorStandard Rigor = "standard"
	RigorCritical Rigor = "critical"
)

type AssessmentStatus string

const (
	AssessmentUnassessed   AssessmentStatus = "unassessed"
	AssessmentInsufficient AssessmentStatus = "insufficient"
	AssessmentAssessed     AssessmentStatus = "assessed"
)

type RequiredPolicyFact struct {
	PolicyID           string       `json:"policyId"`
	Version            string       `json:"version"`
	DefinitionDigest   string       `json:"definitionDigest"`
	Issuer             string       `json:"issuer"`
	Validity           FactValidity `json:"validity"`
	MinimumRigor       Rigor        `json:"minimumRigor"`
	RequiredProfiles   []string     `json:"requiredProfiles"`
	RequiredValidators []string     `json:"requiredValidators"`
	EvidenceRefs       []string     `json:"evidenceRefs"`
	IssuedAt           time.Time    `json:"issuedAt"`
	ExpiresAt          *time.Time   `json:"expiresAt,omitempty"`
}

type ProjectPolicyFact struct {
	PolicyID           string       `json:"policyId"`
	Version            string       `json:"version"`
	DefinitionDigest   string       `json:"definitionDigest"`
	Issuer             string       `json:"issuer"`
	Validity           FactValidity `json:"validity"`
	MinimumRigor       Rigor        `json:"minimumRigor"`
	RequiredProfiles   []string     `json:"requiredProfiles"`
	RequiredValidators []string     `json:"requiredValidators"`
	EvidenceRefs       []string     `json:"evidenceRefs"`
	IssuedAt           time.Time    `json:"issuedAt"`
	ExpiresAt          *time.Time   `json:"expiresAt,omitempty"`
}

type AssessmentFact struct {
	AssessmentID       string           `json:"assessmentId"`
	Status             AssessmentStatus `json:"status"`
	Issuer             string           `json:"issuer"`
	Validity           FactValidity     `json:"validity"`
	RecommendedRigor   Rigor            `json:"recommendedRigor,omitempty"`
	RequiredProfiles   []string         `json:"requiredProfiles"`
	RequiredValidators []string         `json:"requiredValidators"`
	EvidenceRefs       []string         `json:"evidenceRefs"`
	Uncertainty        string           `json:"uncertainty"`
	SourceRevision     uint64           `json:"sourceRevision"`
	ObservedAt         time.Time        `json:"observedAt"`
}

type CompositionRef struct {
	CompositionID      string                `json:"compositionId"`
	Revision           uint64                `json:"revision"`
	Digest             string                `json:"digest"`
	JourneyRef         string                `json:"journeyRef"`
	Rigor              Rigor                 `json:"rigor"`
	Profiles           []string              `json:"profiles"`
	Validators         []string              `json:"validators"`
	GraphDefinitionRef CompiledDefinitionRef `json:"graphDefinitionRef"`
	PolicyDigest       string                `json:"policyDigest"`
}

type CompositionInput struct {
	WorkID                string                `json:"workId"`
	CompositionID         string                `json:"compositionId"`
	ExpectedWorkRevision  uint64                `json:"expectedWorkRevision"`
	PresetRef             string                `json:"presetRef,omitempty"`
	JourneyRef            string                `json:"journeyRef"`
	RequestedRigor        Rigor                 `json:"requestedRigor,omitempty"`
	OrganizationPolicy    RequiredPolicyFact    `json:"organizationPolicy"`
	ProjectPolicy         *ProjectPolicyFact    `json:"projectPolicy,omitempty"`
	WorkRationale         string                `json:"workRationale,omitempty"`
	Assessment            AssessmentFact        `json:"assessment"`
	GraphDefinitionRef    CompiledDefinitionRef `json:"graphDefinitionRef"`
	ArtifactProfileRefs   []string              `json:"artifactProfileRefs"`
	ValidatorContractRefs []string              `json:"validatorContractRefs"`
	Current               *CompositionRef       `json:"current,omitempty"`
	EvaluationInstant     time.Time             `json:"evaluationInstant"`
}

type CompositionStatus string

const (
	CompositionResolved           CompositionStatus = "resolved"
	CompositionBlocked            CompositionStatus = "blocked"
	CompositionEscalationProposed CompositionStatus = "escalation-proposed"
)

type CompositionProposal struct {
	ProposalDigest       string          `json:"proposalDigest"`
	WorkID               string          `json:"workId"`
	ExpectedWorkRevision uint64          `json:"expectedWorkRevision"`
	Current              *CompositionRef `json:"current,omitempty"`
	Proposed             CompositionRef  `json:"proposed"`
	Reasons              []string        `json:"reasons"`
	EvidenceRefs         []string        `json:"evidenceRefs"`
}

type CompositionDecision struct {
	Status       CompositionStatus    `json:"status"`
	Effective    *CompositionRef      `json:"effective,omitempty"`
	Proposal     *CompositionProposal `json:"proposal,omitempty"`
	Reasons      []string             `json:"reasons"`
	EvidenceRefs []string             `json:"evidenceRefs"`
	Diagnostics  []Diagnostic         `json:"diagnostics"`
}

// ResolveComposition applies policy precedence without mutating active work.
// A proposal is not acceptance and acceptance is not workspace application.
func ResolveComposition(ctx context.Context, input CompositionInput) (CompositionDecision, []Diagnostic) {
	if ctx == nil {
		diagnostics := []Diagnostic{diagnostic("WORKFLOW_CONTEXT_INVALID", "composition resolution requires a non-nil context", SourceCoordinate{}, "Pass the request context to workflow.ResolveComposition.")}
		return blockedComposition(diagnostics), diagnostics
	}
	if err := ctx.Err(); err != nil {
		diagnostics := []Diagnostic{cancelledDiagnostic(SourceCoordinate{}, err)}
		return blockedComposition(diagnostics), diagnostics
	}
	if input.Assessment.Status == "" {
		input.Assessment.Status = AssessmentUnassessed
	}
	diagnostics := validateCompositionInput(input)
	if len(diagnostics) != 0 {
		sortDiagnostics(diagnostics)
		return blockedComposition(diagnostics), diagnostics
	}

	effectiveRigor := input.OrganizationPolicy.MinimumRigor
	reasons := []string{"organization-policy-floor"}
	profiles := append([]string(nil), input.OrganizationPolicy.RequiredProfiles...)
	validators := append([]string(nil), input.OrganizationPolicy.RequiredValidators...)
	evidence := append([]string(nil), input.OrganizationPolicy.EvidenceRefs...)

	if input.ProjectPolicy != nil {
		effectiveRigor = maxRigor(effectiveRigor, input.ProjectPolicy.MinimumRigor)
		profiles = append(profiles, input.ProjectPolicy.RequiredProfiles...)
		validators = append(validators, input.ProjectPolicy.RequiredValidators...)
		evidence = append(evidence, input.ProjectPolicy.EvidenceRefs...)
		reasons = append(reasons, "project-policy-narrowing")
	}
	if input.Assessment.Status == AssessmentAssessed {
		effectiveRigor = maxRigor(effectiveRigor, input.Assessment.RecommendedRigor)
		profiles = append(profiles, input.Assessment.RequiredProfiles...)
		validators = append(validators, input.Assessment.RequiredValidators...)
		evidence = append(evidence, input.Assessment.EvidenceRefs...)
		reasons = append(reasons, "trusted-assessment")
	} else {
		reasons = append(reasons, "assessment-"+string(input.Assessment.Status))
	}
	if input.RequestedRigor != "" {
		effectiveRigor = maxRigor(effectiveRigor, input.RequestedRigor)
		reasons = append(reasons, "explicit-request")
	}
	profiles = append(profiles, input.ArtifactProfileRefs...)
	validators = append(validators, input.ValidatorContractRefs...)
	if input.Current != nil {
		if input.Current.JourneyRef != input.JourneyRef {
			diagnostics = []Diagnostic{diagnostic("WORKFLOW_MIGRATION_REQUIRED", "requested journey differs from the active composition", SourceCoordinate{}, "Create and accept an explicit workflow migration proposal.")}
			return blockedComposition(diagnostics), diagnostics
		}
		effectiveRigor = maxRigor(effectiveRigor, input.Current.Rigor)
		profiles = append(profiles, input.Current.Profiles...)
		validators = append(validators, input.Current.Validators...)
	}
	profiles = uniqueSortedIDs(profiles)
	validators = uniqueSortedIDs(validators)
	reasons = uniqueSortedStrings(reasons)
	evidence = uniqueSortedStrings(evidence)

	policyDigest, policyDigestErr := effectivePolicyDigest(input)
	if policyDigestErr != nil {
		diagnostics = []Diagnostic{diagnostic("WORKFLOW_COMPOSITION_INVALID", "effective policy cannot be canonically identified", SourceCoordinate{}, "Report this invariant failure.", MetadataEntry{Key: "cause", Value: policyDigestErr.Error()})}
		return blockedComposition(diagnostics), diagnostics
	}
	resolved := CompositionRef{
		CompositionID:      compositionID(input),
		Revision:           1,
		JourneyRef:         input.JourneyRef,
		Rigor:              effectiveRigor,
		Profiles:           profiles,
		Validators:         validators,
		GraphDefinitionRef: input.GraphDefinitionRef,
		PolicyDigest:       policyDigest,
	}
	if input.Current != nil {
		resolved.Revision = input.Current.Revision
	}
	digest, digestErr := compositionDigest(resolved)
	if digestErr != nil {
		diagnostics = []Diagnostic{diagnostic("WORKFLOW_COMPOSITION_INVALID", "resolved composition cannot be canonically identified", SourceCoordinate{}, "Report this invariant failure.", MetadataEntry{Key: "cause", Value: digestErr.Error()})}
		return blockedComposition(diagnostics), diagnostics
	}
	resolved.Digest = digest

	if input.Current == nil {
		decision := CompositionDecision{Status: CompositionResolved, Effective: &resolved, Reasons: reasons, EvidenceRefs: evidence, Diagnostics: []Diagnostic{}}
		return decision, nil
	}
	if compositionsEquivalent(*input.Current, resolved) {
		effective := cloneCompositionRef(input.Current)
		decision := CompositionDecision{Status: CompositionResolved, Effective: effective, Reasons: reasons, EvidenceRefs: evidence, Diagnostics: []Diagnostic{}}
		return decision, nil
	}
	resolved.Revision = nextCompositionRevision(input.Current)
	resolved.Digest, digestErr = compositionDigest(resolved)
	if digestErr != nil {
		diagnostics = []Diagnostic{diagnostic("WORKFLOW_COMPOSITION_INVALID", "proposed composition cannot be canonically identified", SourceCoordinate{}, "Report this invariant failure.", MetadataEntry{Key: "cause", Value: digestErr.Error()})}
		return blockedComposition(diagnostics), diagnostics
	}
	proposal := CompositionProposal{
		WorkID:               input.WorkID,
		ExpectedWorkRevision: input.ExpectedWorkRevision,
		Current:              cloneCompositionRef(input.Current),
		Proposed:             resolved,
		Reasons:              reasons,
		EvidenceRefs:         evidence,
	}
	proposalDigest, proposalErr := digestCompositionProposal(proposal)
	if proposalErr != nil {
		diagnostics = []Diagnostic{diagnostic("WORKFLOW_COMPOSITION_INVALID", "composition proposal cannot be canonically identified", SourceCoordinate{}, "Report this invariant failure.", MetadataEntry{Key: "cause", Value: proposalErr.Error()})}
		return blockedComposition(diagnostics), diagnostics
	}
	proposal.ProposalDigest = proposalDigest
	decision := CompositionDecision{Status: CompositionEscalationProposed, Effective: cloneCompositionRef(input.Current), Proposal: &proposal, Reasons: reasons, EvidenceRefs: evidence, Diagnostics: []Diagnostic{}}
	return decision, nil
}

func validateCompositionInput(input CompositionInput) []Diagnostic {
	var diagnostics []Diagnostic
	if strings.TrimSpace(input.WorkID) == "" || strings.TrimSpace(input.CompositionID) == "" || !validStableID(input.JourneyRef) || !input.GraphDefinitionRefValid() {
		diagnostics = append(diagnostics, diagnostic("WORKFLOW_COMPOSITION_INVALID", "composition work, composition, journey, or graph identity is invalid", SourceCoordinate{}, "Provide opaque work/composition IDs, a stable journey ID, and an exact compiled graph reference."))
	}
	if input.RequestedRigor != "" && !validRigor(input.RequestedRigor) {
		diagnostics = append(diagnostics, diagnostic("WORKFLOW_COMPOSITION_INVALID", "requested rigor is invalid", SourceCoordinate{}, "Use quick, standard, or critical."))
	}
	diagnostics = append(diagnostics, validateRequiredPolicy(input.OrganizationPolicy, input.EvaluationInstant)...)
	if input.ProjectPolicy != nil {
		diagnostics = append(diagnostics, validateProjectPolicy(*input.ProjectPolicy, input.EvaluationInstant)...)
	}
	diagnostics = append(diagnostics, validateAssessment(input.Assessment, input.EvaluationInstant)...)
	if input.Current != nil {
		if input.Current.CompositionID != input.CompositionID || !validRigor(input.Current.Rigor) || input.Current.Digest == "" {
			diagnostics = append(diagnostics, diagnostic("WORKFLOW_COMPOSITION_INVALID", "active composition reference is invalid", SourceCoordinate{}, "Load the exact active composition reference."))
		}
	}
	for _, value := range append(slices.Clone(input.ArtifactProfileRefs), input.ValidatorContractRefs...) {
		if !validStableID(value) {
			diagnostics = append(diagnostics, diagnostic("WORKFLOW_COMPOSITION_INVALID", "profile or validator reference is invalid", SourceCoordinate{}, "Use stable lowercase profile and validator IDs.", MetadataEntry{Key: "reference", Value: value}))
		}
	}
	return diagnostics
}

func (input CompositionInput) GraphDefinitionRefValid() bool {
	return input.GraphDefinitionRef.CompiledDigest != "" && input.GraphDefinitionRef.CompilerVersion == CompilerContractVersion
}

func validateRequiredPolicy(policy RequiredPolicyFact, instant time.Time) []Diagnostic {
	return validatePolicy("organization", policy.PolicyID, policy.Version, policy.DefinitionDigest, policy.Issuer, policy.Validity, policy.MinimumRigor, policy.RequiredProfiles, policy.RequiredValidators, policy.IssuedAt, policy.ExpiresAt, instant)
}

func validateProjectPolicy(policy ProjectPolicyFact, instant time.Time) []Diagnostic {
	return validatePolicy("project", policy.PolicyID, policy.Version, policy.DefinitionDigest, policy.Issuer, policy.Validity, policy.MinimumRigor, policy.RequiredProfiles, policy.RequiredValidators, policy.IssuedAt, policy.ExpiresAt, instant)
}

func validatePolicy(kind, id, version, digest, issuer string, validity FactValidity, rigor Rigor, profiles, validators []string, issuedAt time.Time, expiresAt *time.Time, instant time.Time) []Diagnostic {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(version) == "" || digest == "" || strings.TrimSpace(issuer) == "" || !validRigor(rigor) {
		return []Diagnostic{diagnostic("WORKFLOW_POLICY_UNAVAILABLE", kind+" policy identity or floor is invalid", SourceCoordinate{}, "Load a current trusted policy fact.")}
	}
	if validity != ValidityCurrent {
		return []Diagnostic{diagnostic("WORKFLOW_POLICY_UNAVAILABLE", kind+" policy is not current", SourceCoordinate{}, "Refresh the required policy fact.", MetadataEntry{Key: "validity", Value: string(validity)})}
	}
	if instant.IsZero() || issuedAt.IsZero() || issuedAt.After(instant) || (expiresAt != nil && !instant.Before(*expiresAt)) {
		return []Diagnostic{diagnostic("WORKFLOW_POLICY_UNAVAILABLE", kind+" policy time window is invalid or expired", SourceCoordinate{}, "Refresh the required policy and evaluation instant.")}
	}
	for _, value := range append(slices.Clone(profiles), validators...) {
		if !validStableID(value) {
			return []Diagnostic{diagnostic("WORKFLOW_POLICY_UNAVAILABLE", kind+" policy contains an invalid profile or validator", SourceCoordinate{}, "Correct the trusted policy fact.")}
		}
	}
	return nil
}

func validateAssessment(assessment AssessmentFact, instant time.Time) []Diagnostic {
	if assessment.Status != AssessmentUnassessed && assessment.Status != AssessmentInsufficient && assessment.Status != AssessmentAssessed {
		return []Diagnostic{diagnostic("WORKFLOW_ASSESSMENT_INVALID", "assessment status is invalid", SourceCoordinate{}, "Use unassessed, insufficient, or assessed.")}
	}
	if assessment.Status != AssessmentAssessed {
		if assessment.RecommendedRigor != "" || len(assessment.RequiredProfiles) != 0 || len(assessment.RequiredValidators) != 0 {
			return []Diagnostic{diagnostic("WORKFLOW_ASSESSMENT_INVALID", "unassessed or insufficient fact asserts authoritative requirements", SourceCoordinate{}, "Remove inferred requirements until a trusted assessment is complete.")}
		}
		return nil
	}
	if strings.TrimSpace(assessment.AssessmentID) == "" || strings.TrimSpace(assessment.Issuer) == "" || assessment.Validity != ValidityCurrent || !validRigor(assessment.RecommendedRigor) || assessment.SourceRevision == 0 || assessment.ObservedAt.IsZero() || instant.IsZero() || assessment.ObservedAt.After(instant) || len(assessment.EvidenceRefs) == 0 || strings.TrimSpace(assessment.Uncertainty) == "" {
		return []Diagnostic{diagnostic("WORKFLOW_ASSESSMENT_INVALID", "assessed fact lacks trusted identity, evidence, revision, or rigor", SourceCoordinate{}, "Provide a complete current assessment fact.")}
	}
	for _, value := range append(slices.Clone(assessment.RequiredProfiles), assessment.RequiredValidators...) {
		if !validStableID(value) {
			return []Diagnostic{diagnostic("WORKFLOW_ASSESSMENT_INVALID", "assessed fact contains an invalid profile or validator", SourceCoordinate{}, "Correct the trusted assessment fact.")}
		}
	}
	return nil
}

func validRigor(value Rigor) bool {
	return value == RigorQuick || value == RigorStandard || value == RigorCritical
}

func maxRigor(left, right Rigor) Rigor {
	if rigorRank(right) > rigorRank(left) {
		return right
	}
	return left
}

func rigorRank(value Rigor) int {
	switch value {
	case RigorQuick:
		return 1
	case RigorStandard:
		return 2
	case RigorCritical:
		return 3
	default:
		return 0
	}
}

func uniqueSortedIDs(values []string) []string { return uniqueSortedStrings(values) }

func uniqueSortedStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return slices.Compact(result)
}

func compositionID(input CompositionInput) string {
	return input.CompositionID
}

func nextCompositionRevision(current *CompositionRef) uint64 {
	if current == nil {
		return 1
	}
	return current.Revision + 1
}

func compositionDigest(value CompositionRef) (string, error) {
	value.Digest = ""
	encoded, err := canonicaljson.Marshal(value)
	if err != nil {
		return "", err
	}
	return digestWithDomain("sdd.workflow.composition.v1\x00", encoded), nil
}

func effectivePolicyDigest(input CompositionInput) (string, error) {
	organization := input.OrganizationPolicy
	organization.RequiredProfiles = uniqueSortedIDs(organization.RequiredProfiles)
	organization.RequiredValidators = uniqueSortedIDs(organization.RequiredValidators)
	organization.EvidenceRefs = uniqueSortedStrings(organization.EvidenceRefs)
	organization.IssuedAt = organization.IssuedAt.UTC()
	if organization.ExpiresAt != nil {
		expiresAt := organization.ExpiresAt.UTC()
		organization.ExpiresAt = &expiresAt
	}
	var project *ProjectPolicyFact
	if input.ProjectPolicy != nil {
		copyValue := *input.ProjectPolicy
		copyValue.RequiredProfiles = uniqueSortedIDs(copyValue.RequiredProfiles)
		copyValue.RequiredValidators = uniqueSortedIDs(copyValue.RequiredValidators)
		copyValue.EvidenceRefs = uniqueSortedStrings(copyValue.EvidenceRefs)
		copyValue.IssuedAt = copyValue.IssuedAt.UTC()
		if copyValue.ExpiresAt != nil {
			expiresAt := copyValue.ExpiresAt.UTC()
			copyValue.ExpiresAt = &expiresAt
		}
		project = &copyValue
	}
	projection := struct {
		Organization RequiredPolicyFact `json:"organization"`
		Project      *ProjectPolicyFact `json:"project,omitempty"`
	}{Organization: organization, Project: project}
	encoded, err := canonicaljson.Marshal(projection)
	if err != nil {
		return "", err
	}
	return digestWithDomain("sdd.workflow.effective-policy.v1\x00", encoded), nil
}

func compositionsEquivalent(left, right CompositionRef) bool {
	leftProfiles := uniqueSortedIDs(left.Profiles)
	rightProfiles := uniqueSortedIDs(right.Profiles)
	leftValidators := uniqueSortedIDs(left.Validators)
	rightValidators := uniqueSortedIDs(right.Validators)
	return left.CompositionID == right.CompositionID &&
		left.JourneyRef == right.JourneyRef &&
		left.Rigor == right.Rigor &&
		slices.Equal(leftProfiles, rightProfiles) &&
		slices.Equal(leftValidators, rightValidators) &&
		left.GraphDefinitionRef.CompiledDigest == right.GraphDefinitionRef.CompiledDigest &&
		left.PolicyDigest == right.PolicyDigest
}

func digestCompositionProposal(value CompositionProposal) (string, error) {
	value.ProposalDigest = ""
	encoded, err := canonicaljson.Marshal(value)
	if err != nil {
		return "", err
	}
	return digestWithDomain("sdd.workflow.composition-proposal.v1\x00", encoded), nil
}

func cloneCompositionRef(value *CompositionRef) *CompositionRef {
	if value == nil {
		return nil
	}
	result := *value
	result.Profiles = slices.Clone(value.Profiles)
	result.Validators = slices.Clone(value.Validators)
	return &result
}

func blockedComposition(diagnostics []Diagnostic) CompositionDecision {
	return CompositionDecision{Status: CompositionBlocked, Reasons: []string{}, EvidenceRefs: []string{}, Diagnostics: slices.Clone(diagnostics)}
}
