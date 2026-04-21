package internal

// --- Parser output ---

type ParsedSignals struct {
	Services  []string `json:"services"`
	Endpoints []string `json:"endpoints"`
	Keywords  []string `json:"keywords"`
	Signals   []string `json:"signals"`
	TimeHints []string `json:"time_hints"`
}

// --- Output contract ---

type Category string

const (
	CategoryExternalPaymentProvider Category = "external_payment_provider_issue"
	CategoryDBDegradation           Category = "db_degradation_caused_by_reporting"
	CategoryNotificationDelivery    Category = "notification_delivery_issue"
	CategoryUserAuthentication      Category = "user_authentication_errors"
	CategoryUnknown                 Category = "unknown_or_mixed"
)

var ValidCategories = []Category{
	CategoryExternalPaymentProvider,
	CategoryDBDegradation,
	CategoryNotificationDelivery,
	CategoryUserAuthentication,
	CategoryUnknown,
}

type Severity string

const (
	SeverityLow    Severity = "low"
	SeverityMedium Severity = "medium"
	SeverityHigh   Severity = "high"
)

var ValidSeverities = []Severity{
	SeverityLow,
	SeverityMedium,
	SeverityHigh,
}

type Hypothesis struct {
	Title     string   `json:"title"`
	Reasoning string   `json:"reasoning"`
	NextSteps []string `json:"next_steps"`
}

type TriageResult struct {
	Category   Category     `json:"category"`
	Summary    string       `json:"summary"`
	Affected   string       `json:"affected"`
	Severity   Severity     `json:"severity"`
	Hypotheses []Hypothesis `json:"hypotheses"`
}

// --- Config ---

type Config struct {
	APIKey  string
	Model   string
	Verbose bool
}

// --- Knowledge types ---

type ServiceInfo struct {
	Name                 string   `json:"name"`
	Type                 string   `json:"type"`
	Responsibilities     []string `json:"responsibilities"`
	DependsOn            []string `json:"depends_on"`
	DataStores           []string `json:"data_stores,omitempty"`
	ExternalDependencies []string `json:"external_dependencies,omitempty"`
	KnownRisks           []string `json:"known_risks"`
}

type DatabaseInfo struct {
	Engine       string   `json:"engine"`
	InstanceName string   `json:"instance_name"`
	UsedBy       []string `json:"used_by"`
}

type SystemDescription struct {
	PlatformName         string        `json:"platform_name"`
	Summary              string        `json:"summary"`
	Services             []ServiceInfo `json:"services"`
	SharedInfrastructure struct {
		Logging struct {
			Type   string `json:"type"`
			Stack  string `json:"stack"`
			UsedBy string `json:"used_by"`
		} `json:"logging"`
		Databases []DatabaseInfo `json:"databases"`
	} `json:"shared_infrastructure"`
	KnownOperationalPatterns []struct {
		Pattern      string   `json:"pattern"`
		Signals      []string `json:"signals"`
		LikelyImpact string   `json:"likely_impact"`
	} `json:"known_operational_patterns"`
}

type PastIncident struct {
	ID              string   `json:"id"`
	Category        string   `json:"category"`
	Title           string   `json:"title"`
	Summary         string   `json:"summary"`
	Services        []string `json:"services"`
	Endpoints       []string `json:"endpoints"`
	Keywords        []string `json:"keywords"`
	Signals         []string `json:"signals"`
	LikelyAffected  string   `json:"likely_affected"`
	DiagnosticHints []string `json:"diagnostic_hints"`
}
