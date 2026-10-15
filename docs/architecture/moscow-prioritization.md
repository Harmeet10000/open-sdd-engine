# MoSCoW scope-priority contract

## Purpose

`sdd` uses MoSCoW to negotiate a named delivery scope without confusing product
priority with technical readiness, risk, severity, authority, or task
dependency order. It is a profile-governed planning contract used most directly
by Quick Spec and consumed by completion evidence.

The method originates with Dai Clegg's 1994 DSDM practice. This document does
not treat that history as a license for loose labels: in `sdd`, every category
is a reviewable statement about one pinned delivery commitment.

The normative behavior is in `sdd-guided-workflows`,
`sdd-artifact-validation`, and `sdd-completion-lifecycle`.

## The four dispositions

| Value | Meaning for the delivery scope | Required record |
|---|---|---|
| `must` | The scope cannot be accepted/released without it. | Non-negotiable rationale and why no acceptable time-bounded workaround exists. |
| `should` | High value, but a bounded workaround or deferral exists. | Workaround, owner, operating limit, and reconsideration window. |
| `could` | Optional scope that may be dropped without invalidating the commitment. | Benefit and selection condition. |
| `wont` | Recognized but excluded from this delivery scope; not “never.” | Exclusion rationale and reconsideration trigger/release. |

Categories are local to a pinned scope baseline. A requirement can be a `must`
for one regulated release and a `should` for an internal prototype, but cannot
silently change because an agent edits prose or a capacity guess changes.
Each stable subject appears once in a baseline. Hierarchy does not implicitly
inherit a category: a parent `must` cannot turn every child task into a Must,
and the same subject cannot be counted as both Must and Could.

## Decision and capacity rule

`must` is an acceptance-floor commitment, not a stakeholder preference. The
decision record answers whether the core workflow, an explicit
safety/compliance obligation, or a declared operational release condition fails
without the item; whether an acceptable bounded workaround exists; what
observable harm deferral causes; and which authority accepted the baseline.

An acceptable workaround normally means `should`. The exception must cite why
that workaround is unacceptable. A `must` can never be inferred from a label
in agent-generated prose.

When a delivery scope has comparable estimates and declared usable capacity,
the default profile caps `must` work at **60%**. The baseline pins its estimate
unit or conversion rule and prevents a shared task from being counted twice.
`should` and `could` are planning buffers, commonly targeted near 20% each,
not promises to consume all remaining capacity. Policy may choose a different
documented cap, but missing estimates or unknown capacity never count as a
passing budget.

If must work exceeds the cap, `sdd` returns a scope-replan action. It does not
relabel work automatically. Under capacity loss, explicitly revise the baseline
by dropping `could` first and `should` second. A revision binds the predecessor
digest and expected scope revision. Changing a `must` requires the authority
that controls the delivery scope; it is never silently demoted to make a
dashboard green.

## Execution, rollout, and scoring boundaries

MoSCoW does not decide whether an action is technically ready, authorized, or
safe to apply. Dependencies, validation, policy, and completion gates remain
authoritative for those decisions.

A feature flag is rollout evidence, not a MoSCoW category. A `should` or
`could` feature may be delivered behind a flag and disabled under a declared
latency/error budget. An automatic rule must name the measured signal,
evaluation window, eligible cohort, fallback, observability, owner, and
removal/review condition; it changes rollout availability, not the signed
scope baseline. A flag cannot make an unimplemented `must` complete or hide a
safety/compliance gate.

RICE, WSJF, and similar scores may be advisory evidence. They do not
mechanically map to MoSCoW because a rank is not an acceptance-floor decision.

## Minimal record shape

```text
ScopePriority {
  schemaVersion,
  subjectRef, deliveryScopeRef,
  category: must | should | could | wont,
  rationale, decisionEvidenceRefs[],
  workaround?, deferredTo?,
  estimate?, capacityUnit?,
  policyRef, baselineDigest,
  decidedBy, decidedAt
}
```

The canonical form includes every decision-bearing field. Display labels and
incidental timestamps do not change scope meaning; the immutable record retains
them for audit.

## Completion rule

Every applicable `must` must be satisfied, or be changed through an explicit
authorized scope-baseline revision, before completion. A `should` or `could`
can be deferred only with a recorded disposition, workaround/follow-up, and
baseline. A `wont` item is neither an unfinished task nor delivered scope. If
an excluded item was intentionally implemented, it is a scope change that must
be surfaced and re-baselined; it is not uncredited “extra” work.

DONE reports all four categories separately so scope reduction is visible
without being misreported as a verification failure.

## For the chosen ones

MoSCoW is most useful as a *commitment boundary*, not as a sorted backlog.
Ranking says what you want first; a scope baseline says what makes a specific
release claim true. Bind categories to that baseline digest and a category can
change—but only by changing the release claim itself.
