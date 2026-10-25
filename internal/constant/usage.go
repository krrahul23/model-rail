package constant

// UsageSource records where token counts came from. Reported usage is what the
// provider returned; estimated usage is computed locally when the provider's
// final usage chunk was unavailable (for example, a broken stream).
type UsageSource string

// UsageSource constants
const (
	// UsageReported means the provider supplied the final usage numbers.
	UsageReported UsageSource = "reported"
	// UsageEstimated means ModelRail computed the numbers locally.
	UsageEstimated UsageSource = "estimated"
)
