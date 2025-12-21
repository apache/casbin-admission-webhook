package webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/casbin/casbin/v3"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// Server represents the admission webhook server
type Server struct {
	enforcer *casbin.Enforcer
}

// NewServer creates a new webhook server
func NewServer(enforcer *casbin.Enforcer) *Server {
	return &Server{
		enforcer: enforcer,
	}
}

// HandleAdmission handles admission review requests
func (s *Server) HandleAdmission(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		klog.Errorf("failed to read request body: %v", err)
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse the admission review request
	admissionReview := admissionv1.AdmissionReview{}
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		klog.Errorf("failed to unmarshal admission review: %v", err)
		http.Error(w, "failed to unmarshal admission review", http.StatusBadRequest)
		return
	}

	// Create response
	response := s.admit(admissionReview.Request)
	admissionReview.Response = response

	// Write response
	responseBytes, err := json.Marshal(admissionReview)
	if err != nil {
		klog.Errorf("failed to marshal response: %v", err)
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(responseBytes); err != nil {
		klog.Errorf("failed to write response: %v", err)
	}
}

// admit processes the admission request and returns a response
func (s *Server) admit(request *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	response := &admissionv1.AdmissionResponse{
		UID: request.UID,
	}

	// Extract user information
	user := request.UserInfo.Username
	if user == "" {
		user = "system:anonymous"
	}

	// Extract resource information
	resource := request.Resource.Resource
	namespace := request.Namespace
	operation := string(request.Operation)

	klog.V(2).Infof("admission review for user=%s, resource=%s, namespace=%s, operation=%s",
		user, resource, namespace, operation)

	// Build the Casbin request
	// Format: user, resource, namespace, operation
	subject := user
	object := fmt.Sprintf("%s/%s", resource, namespace)
	action := operation

	// Check the policy
	allowed, err := s.enforcer.Enforce(subject, object, action)
	if err != nil {
		klog.Errorf("failed to enforce policy: %v", err)
		response.Allowed = false
		response.Result = &metav1.Status{
			Status:  "Failure",
			Message: fmt.Sprintf("policy enforcement error: %v", err),
			Code:    http.StatusInternalServerError,
		}
		return response
	}

	response.Allowed = allowed
	if !allowed {
		response.Result = &metav1.Status{
			Status:  "Failure",
			Message: fmt.Sprintf("denied by policy: user '%s' cannot '%s' on '%s'", user, action, object),
			Code:    http.StatusForbidden,
		}
		klog.V(2).Infof("admission denied: user=%s, resource=%s, namespace=%s, operation=%s",
			user, resource, namespace, operation)
	} else {
		klog.V(2).Infof("admission allowed: user=%s, resource=%s, namespace=%s, operation=%s",
			user, resource, namespace, operation)
	}

	return response
}

// HandleHealth handles health check requests
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
