package httpapi

import "testing"

func TestSafeNext(t *testing.T) {
	allowed := []string{"/app", "/app/studio/projects", "/invite?token=abc"}
	for _, value := range allowed { if got:=safeNext(value);got!=value{t.Fatalf("safeNext(%q)=%q",value,got)} }
	blocked := []string{"https://evil.example", "//evil.example", "/auth/email-link", "/invite"}
	for _, value := range blocked { if got:=safeNext(value);got!="/app"{t.Fatalf("safeNext(%q)=%q",value,got)} }
}

func TestTokenHasMFA(t *testing.T) {
	if tokenHasMFA(map[string]any{}) { t.Fatal("empty claims must not have MFA") }
	claims:=map[string]any{"firebase":map[string]any{"sign_in_second_factor":"totp-id"}}
	if !tokenHasMFA(claims) { t.Fatal("second-factor claim should be recognized") }
}
