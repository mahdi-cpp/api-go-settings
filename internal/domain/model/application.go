package model

import "time"

func (a *Application) SetID(id int)                    { a.ID = id }
func (a *Application) SetCreationDate(t time.Time)     { a.CreationDate = t }
func (a *Application) SetModificationDate(t time.Time) { a.ModificationDate = t }
func (a *Application) GetID() int                      { return a.ID }
func (a *Application) GetCreationDate() time.Time      { return a.CreationDate }
func (a *Application) GetModificationDate() time.Time  { return a.ModificationDate }

type Application struct {
	ID               int       `json:"id"`
	Name             string    `json:"title"`
	Subtitle         string    `json:"subtitle"`
	Count            int       `json:"count"`
	IsCollection     bool      `json:"isCollection"`
	IsHidden         bool      `json:"isHidden"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
}

// APKInfo is the main struct for holding all information related to an APK file.
type APKInfo struct {
	// --- Basic Parameters (Essential) ---
	PackageName      string     `json:"packageName"`                // Unique identifier for each Android application (e.g., com.example.app)
	VersionCode      int        `json:"versionCode"`                // Internal version code of the application
	VersionName      string     `json:"versionName"`                // Human-readable version name for the user (e.g., 1.0.0)
	FileSize         int64      `json:"fileSize"`                   // Size of the APK file in bytes
	MD5Checksum      string     `json:"md5Checksum"`                // MD5 Hash of the file
	SHA1Checksum     string     `json:"sha1Checksum"`               // SHA-1 Hash of the file
	SHA256Checksum   string     `json:"sha256Checksum"`             // SHA-256 Hash of the file
	InstallationDate *time.Time `json:"installationDate,omitempty"` // Date of installation or last update (optional)
	FilePath         string     `json:"filePath"`                   // Path where the APK file is stored on the system

	// --- Supplementary Parameters (Useful) ---
	ApplicationName    string    `json:"applicationName"`    // User-visible name of the application
	IconPath           string    `json:"iconPath,omitempty"` // Path or file name of the application icon (optional)
	MinSDKVersion      int       `json:"minSDKVersion"`      // Minimum SDK version required
	TargetSDKVersion   int       `json:"targetSDKVersion"`   // Target SDK version
	Permissions        []string  `json:"permissions"`        // List of permissions requested
	Activities         []string  `json:"activities"`         // List of activities defined in the Manifest
	Services           []string  `json:"services"`           // List of services defined in the Manifest
	Receivers          []string  `json:"receivers"`          // List of broadcast receivers defined in the Manifest
	Providers          []string  `json:"providers"`          // List of content providers defined in the Manifest
	Signature          Signature `json:"signature"`          // APK signature information
	SupportedABIs      []string  `json:"supportedABIs"`      // Supported processor architectures (ABIs)
	SupportedLanguages []string  `json:"supportedLanguages"` // Supported languages

	// --- Advanced/Analytical Parameters (Optional) ---
	DiscoveredURLs   []string `json:"discoveredURLs,omitempty"`   // URLs found within the APK's code or resources
	DiscoveredIPs    []string `json:"discoveredIPs,omitempty"`    // IP addresses found within the APK's code or resources
	HardcodedStrings []string `json:"hardcodedStrings,omitempty"` // Important hardcoded strings (e.g., API keys, passwords)
	Developer        string   `json:"developer,omitempty"`        // Developer name (optional)
	Category         string   `json:"category,omitempty"`         // Application category (optional)
	Description      string   `json:"description,omitempty"`      // Textual description of the application (optional)
}

// Signature struct for holding APK signature-related information.
type Signature struct {
	SignerName        string     `json:"signerName"`           // Name of the signer/publisher
	FingerprintMD5    string     `json:"fingerprintMD5"`       // MD5 fingerprint of the certificate
	FingerprintSHA1   string     `json:"fingerprintSHA1"`      // SHA-1 fingerprint of the certificate
	FingerprintSHA256 string     `json:"fingerprintSHA256"`    // SHA-256 fingerprint of the certificate
	ValidFrom         *time.Time `json:"validFrom,omitempty"`  // Certificate valid from date
	ValidUntil        *time.Time `json:"validUntil,omitempty"` // Certificate valid until date
}

// Additional structs could be defined for more detailed internal file or class information if needed.
// Example:
// type APKFile struct {
// 	Name string `json:"name"`
// 	Size int64  `json:"size"`
// 	Path string `json:"path"`
// 	Type string `json:"type"` // e.g., "dex", "resource", "lib"
// }

// type ClassInfo struct {
// 	Name    string   `json:"name"`
// 	Methods []string `json:"methods"`
// }
