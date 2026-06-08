package signature

import "testing"

func TestVerifierVerify(t *testing.T) {
	verifier, err := NewVerifier(Config{
		Enabled: true,
		Secret:  "demo-secret",
	})
	if err != nil {
		t.Fatalf("NewVerifier() error = %v", err)
	}

	sig := verifier.Sign("GET", "/console/v1/tasks/1001")
	if err := verifier.Verify("GET", "/console/v1/tasks/1001", sig); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if err := verifier.Verify("GET", "/console/v1/tasks/1001", "bad-signature"); err == nil {
		t.Fatalf("Verify() expected invalid signature error")
	}
}

func TestVerifierDisabled(t *testing.T) {
	verifier, err := NewVerifier(Config{Enabled: false})
	if err != nil {
		t.Fatalf("NewVerifier() error = %v", err)
	}
	if err := verifier.Verify("GET", "/console/v1/tasks/1001", ""); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
}
