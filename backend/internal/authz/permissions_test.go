package authz

import "testing"

func TestRolePermissions(t *testing.T) {
	if !Allows(Superadmin, "tenant:create") {
		t.Fatal("superadmin should create tenants")
	}
	if Allows(Viewer, "file:delete") {
		t.Fatal("viewer must not delete files")
	}
	if Allows(Staff, "platform:manage") {
		t.Fatal("staff must not manage platform")
	}
}
