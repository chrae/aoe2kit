package kit

// An archive has to be able to say what it is. Without that, verification can
// only hold every copy to the fullest possible contract, so a deliberately
// reduced archive reads as a corrupt one — and a recipient who cannot verify
// cannot pack, which means they cannot pass the kit on. The marker is what
// makes a reduced archive valid on its own terms.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const ProfileMarkerName = "KIT_PROFILE.json"

type ProfileMarker struct {
	Profile    string `json:"profile"`
	KitVersion string `json:"kit_version"`
	PackedAt   string `json:"packed_at"`
	// Generation counts how many times this kit has been packed and repacked.
	// A copy three hands down can say so without any central registry.
	Generation int    `json:"generation"`
	ParentSHA  string `json:"parent_sha256,omitempty"`
}

// KnownProfile reports whether a declared profile is one this kit understands.
func KnownProfile(p PackProfile) bool {
	switch p {
	case ProfileFull, ProfileHandoff, ProfileSandbox, ProfilePublic:
		return true
	}
	return false
}

// ReadProfileMarker returns the archive's declared profile. A tree with no
// marker is a full kit: that is what every existing checkout is. An unreadable
// or unrecognized marker is reported rather than quietly treated as full — on a
// redistribution gate, "I do not know what this archive claims to be" must not
// look like "this archive is fine".
func ReadProfileMarker(root string) (ProfileMarker, error) {
	marker := ProfileMarker{Profile: string(ProfileFull)}
	data, err := os.ReadFile(filepath.Join(root, ProfileMarkerName))
	if err != nil {
		if os.IsNotExist(err) {
			return marker, nil
		}
		return marker, fmt.Errorf("read %s: %w", ProfileMarkerName, err)
	}
	var parsed ProfileMarker
	if err := json.Unmarshal(data, &parsed); err != nil {
		return marker, fmt.Errorf("%s is not valid JSON: %w", ProfileMarkerName, err)
	}
	if parsed.Profile == "" {
		return marker, fmt.Errorf("%s declares no profile", ProfileMarkerName)
	}
	if !KnownProfile(PackProfile(parsed.Profile)) {
		return parsed, fmt.Errorf("unknown KIT_PROFILE profile %q (this kit understands full, handoff, sandbox, public)", parsed.Profile)
	}
	return parsed, nil
}

// nextMarker stamps the archive being produced, carrying lineage forward from
// the tree it was packed from.
func nextMarker(root string, profile PackProfile, now time.Time) ProfileMarker {
	parent, _ := ReadProfileMarker(root)
	return ProfileMarker{
		Profile:    string(profile),
		KitVersion: Version,
		PackedAt:   now.UTC().Format(time.RFC3339),
		Generation: parent.Generation + 1,
		ParentSHA:  parent.ParentSHA,
	}
}

// RequiredDocs returns the artifacts a given profile promises to carry. It is
// the full list minus what the profile excludes, so the promise and the packing
// rules can never disagree.
func RequiredDocs(profile PackProfile) []string {
	excludes := profile.excludes()
	var required []string
	for _, doc := range CurrentManifest().Docs {
		if excluded(strings.TrimPrefix(doc, "./"), excludes) {
			continue
		}
		required = append(required, doc)
	}
	return required
}
