package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/casbin/casbin/v2"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createTestEnforcer(t *testing.T) *casbin.Enforcer {
	tmpDir := t.TempDir()

	// Create test model
	modelFile := filepath.Join(tmpDir, "model.conf")
	modelContent := `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && (r.obj == p.obj || p.obj == "*") && (r.act == p.act || p.act == "*")
`
	if err := os.WriteFile(modelFile, []byte(modelContent), 0644); err != nil {
		t.Fatalf("Failed to create model file: %v", err)
	}

	// Create test policy
	policyFile := filepath.Join(tmpDir, "policy.csv")
	policyContent := `p, admin, *, *
p, developer, pods/development, CREATE
p, viewer, */*, GET
`
	if err := os.WriteFile(policyFile, []byte(policyContent), 0644); err != nil {
		t.Fatalf("Failed to create policy file: %v", err)
	}

	enforcer, err := casbin.NewEnforcer(modelFile, policyFile)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	return enforcer
}

func TestHandleAdmission(t *testing.T) {
	enforcer := createTestEnforcer(t)
	server := NewServer(enforcer)

	tests := []struct {
		name           string
		request        *admissionv1.AdmissionRequest
		expectedAllow  bool
		expectedStatus int
	}{
		{
			name: "Admin can do anything",
			request: &admissionv1.AdmissionRequest{
				UID: "test-uid-1",
				UserInfo: authenticationv1.UserInfo{
					Username: "admin",
				},
				Resource: metav1.GroupVersionResource{
					Resource: "pods",
				},
				Namespace: "default",
				Operation: admissionv1.Create,
			},
			expectedAllow:  true,
			expectedStatus: 0,
		},
		{
			name: "Developer can create pods in development namespace",
			request: &admissionv1.AdmissionRequest{
				UID: "test-uid-2",
				UserInfo: authenticationv1.UserInfo{
					Username: "developer",
				},
				Resource: metav1.GroupVersionResource{
					Resource: "pods",
				},
				Namespace: "development",
				Operation: admissionv1.Create,
			},
			expectedAllow:  true,
			expectedStatus: 0,
		},
		{
			name: "Developer cannot create pods in production namespace",
			request: &admissionv1.AdmissionRequest{
				UID: "test-uid-3",
				UserInfo: authenticationv1.UserInfo{
					Username: "developer",
				},
				Resource: metav1.GroupVersionResource{
					Resource: "pods",
				},
				Namespace: "production",
				Operation: admissionv1.Create,
			},
			expectedAllow:  false,
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Viewer can get resources",
			request: &admissionv1.AdmissionRequest{
				UID: "test-uid-4",
				UserInfo: authenticationv1.UserInfo{
					Username: "viewer",
				},
				Resource: metav1.GroupVersionResource{
					Resource: "pods",
				},
				Namespace: "default",
				Operation: admissionv1.Connect, // Using Connect as a proxy for GET
			},
			expectedAllow:  false,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			admissionReview := admissionv1.AdmissionReview{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "admission.k8s.io/v1",
					Kind:       "AdmissionReview",
				},
				Request: tt.request,
			}

			body, err := json.Marshal(admissionReview)
			if err != nil {
				t.Fatalf("Failed to marshal admission review: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			server.HandleAdmission(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}

			var responseReview admissionv1.AdmissionReview
			if err := json.NewDecoder(resp.Body).Decode(&responseReview); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if responseReview.Response.Allowed != tt.expectedAllow {
				t.Errorf("Expected allowed=%v, got %v", tt.expectedAllow, responseReview.Response.Allowed)
			}

			if !tt.expectedAllow && responseReview.Response.Result != nil {
				if responseReview.Response.Result.Code != int32(tt.expectedStatus) {
					t.Errorf("Expected status code %d, got %d", tt.expectedStatus, responseReview.Response.Result.Code)
				}
			}
		})
	}
}

func TestHandleHealth(t *testing.T) {
	enforcer := createTestEnforcer(t)
	server := NewServer(enforcer)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.HandleHealth(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestHandleAdmissionInvalidMethod(t *testing.T) {
	enforcer := createTestEnforcer(t)
	server := NewServer(enforcer)

	req := httptest.NewRequest(http.MethodGet, "/validate", nil)
	w := httptest.NewRecorder()

	server.HandleAdmission(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}
}

func TestHandleAdmissionInvalidJSON(t *testing.T) {
	enforcer := createTestEnforcer(t)
	server := NewServer(enforcer)

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	server.HandleAdmission(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestAdmitWithEmptyUsername(t *testing.T) {
	enforcer := createTestEnforcer(t)
	server := NewServer(enforcer)

	request := &admissionv1.AdmissionRequest{
		UID: "test-uid",
		UserInfo: authenticationv1.UserInfo{
			Username: "",
		},
		Resource: metav1.GroupVersionResource{
			Resource: "pods",
		},
		Namespace: "default",
		Operation: admissionv1.Create,
	}

	response := server.admit(request)

	// Should default to system:anonymous
	if response.Allowed {
		t.Error("Expected request to be denied for anonymous user")
	}
}
