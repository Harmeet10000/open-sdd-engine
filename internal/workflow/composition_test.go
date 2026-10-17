package workflow

import (
	"context"
	"slices"
	"testing"
	"time"
)

func TestResolveCompositionEnforcesFloorsAndUnionsRequirements(t *testing.T) {
	t.Parallel()
	graph := mustCompile(t, validDefinition())
	input := validCompositionInput(graph)
	input.RequestedRigor = RigorQuick
	input.ProjectPolicy = &ProjectPolicyFact{
		PolicyID:           "project-policy",
		Version:            "1",
		DefinitionDigest:   testDigest,
		Issuer:             "project-authority",
		Validity:           ValidityCurrent,
		MinimumRigor:       RigorCritical,
		RequiredProfiles:   []string{"migration", "regression"},
		RequiredValidators: []string{"integration"},
		EvidenceRefs:       []string{"project-policy-evidence"},
		IssuedAt:           input.EvaluationInstant.Add(-time.Hour),
	}
	input.Assessment = AssessmentFact{
		AssessmentID:       "assessment-1",
		Status:             AssessmentAssessed,
		Issuer:             "risk-authority",
		Validity:           ValidityCurrent,
		RecommendedRigor:   RigorStandard,
		RequiredProfiles:   []string{"regression", "threat-model"},
		RequiredValidators: []string{"security"},
		EvidenceRefs:       []string{"assessment-evidence"},
		Uncertainty:        "low",
		SourceRevision:     4,
		ObservedAt:         input.EvaluationInstant.Add(-time.Minute),
	}

	decision, diagnostics := ResolveComposition(context.Background(), input)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if decision.Status != CompositionResolved || decision.Effective == nil {
		t.Fatalf("decision = %#v", decision)
	}
	if decision.Effective.Rigor != RigorCritical {
		t.Fatalf("rigor = %q, want critical", decision.Effective.Rigor)
	}
	wantProfiles := []string{"migration", "regression", "threat-model"}
	if !slices.Equal(decision.Effective.Profiles, wantProfiles) {
		t.Fatalf("profiles = %#v, want %#v", decision.Effective.Profiles, wantProfiles)
	}
	wantValidators := []string{"integration", "security", "structural"}
	if !slices.Equal(decision.Effective.Validators, wantValidators) {
		t.Fatalf("validators = %#v, want %#v", decision.Effective.Validators, wantValidators)
	}
}

func TestResolveCompositionTreatsMissingAssessmentAsUnassessed(t *testing.T) {
	t.Parallel()
	graph := mustCompile(t, validDefinition())
	input := validCompositionInput(graph)
	input.RequestedRigor = RigorQuick

	decision, diagnostics := ResolveComposition(context.Background(), input)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if decision.Effective == nil || decision.Effective.Rigor != RigorStandard {
		t.Fatalf("decision = %#v", decision)
	}
	if !slices.Contains(decision.Reasons, "assessment-unassessed") {
		t.Fatalf("reasons = %#v", decision.Reasons)
	}
}

func TestResolveCompositionProposesButDoesNotApplyEscalation(t *testing.T) {
	t.Parallel()
	graph := mustCompile(t, validDefinition())
	input := validCompositionInput(graph)
	initial, diagnostics := ResolveComposition(context.Background(), input)
	if len(diagnostics) != 0 || initial.Effective == nil {
		t.Fatalf("initial = %#v diagnostics = %#v", initial, diagnostics)
	}
	current := *initial.Effective
	input.Current = &current
	input.OrganizationPolicy.MinimumRigor = RigorCritical

	decision, diagnostics := ResolveComposition(context.Background(), input)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if decision.Status != CompositionEscalationProposed || decision.Proposal == nil {
		t.Fatalf("decision = %#v", decision)
	}
	if decision.Effective == nil || decision.Effective.Digest != current.Digest {
		t.Fatal("proposal changed the active composition")
	}
	if decision.Proposal.Proposed.Rigor != RigorCritical || decision.Proposal.Proposed.Revision != current.Revision+1 || decision.Proposal.ProposalDigest == "" {
		t.Fatalf("proposal = %#v", decision.Proposal)
	}
}

func TestResolveCompositionFailsClosedForExpiredPolicyAndJourneyChange(t *testing.T) {
	t.Parallel()
	graph := mustCompile(t, validDefinition())
	input := validCompositionInput(graph)
	expired := input.EvaluationInstant.Add(-time.Second)
	input.OrganizationPolicy.ExpiresAt = &expired
	decision, diagnostics := ResolveComposition(context.Background(), input)
	if decision.Status != CompositionBlocked {
		t.Fatalf("expired decision = %#v", decision)
	}
	assertDiagnostic(t, diagnostics, "WORKFLOW_POLICY_UNAVAILABLE")

	input = validCompositionInput(graph)
	initial, diagnostics := ResolveComposition(context.Background(), input)
	if len(diagnostics) != 0 || initial.Effective == nil {
		t.Fatalf("initial = %#v diagnostics = %#v", initial, diagnostics)
	}
	input.Current = initial.Effective
	input.JourneyRef = "bug-fix"
	decision, diagnostics = ResolveComposition(context.Background(), input)
	if decision.Status != CompositionBlocked {
		t.Fatalf("journey-change decision = %#v", decision)
	}
	assertDiagnostic(t, diagnostics, "WORKFLOW_MIGRATION_REQUIRED")
}

func TestResolveCompositionHonorsCancellation(t *testing.T) {
	t.Parallel()
	graph := mustCompile(t, validDefinition())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	decision, diagnostics := ResolveComposition(ctx, validCompositionInput(graph))
	if decision.Status != CompositionBlocked {
		t.Fatalf("decision = %#v", decision)
	}
	assertDiagnostic(t, diagnostics, "WORKFLOW_CANCELLED")
}

func validCompositionInput(graph CompiledGraph) CompositionInput {
	now := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	return CompositionInput{
		WorkID:               "work-1",
		CompositionID:        "composition-1",
		ExpectedWorkRevision: 7,
		PresetRef:            "standard",
		JourneyRef:           "standard-change",
		OrganizationPolicy: RequiredPolicyFact{
			PolicyID:           "organization-policy",
			Version:            "1",
			DefinitionDigest:   testDigest,
			Issuer:             "organization-authority",
			Validity:           ValidityCurrent,
			MinimumRigor:       RigorStandard,
			RequiredProfiles:   []string{},
			RequiredValidators: []string{"structural"},
			EvidenceRefs:       []string{"organization-policy-evidence"},
			IssuedAt:           now.Add(-time.Hour),
		},
		GraphDefinitionRef:    graph.Ref(),
		ArtifactProfileRefs:   []string{},
		ValidatorContractRefs: []string{},
		EvaluationInstant:     now,
	}
}
