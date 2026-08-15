package router_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"wedding_api/ent"
	"wedding_api/internal/auth"
	"wedding_api/internal/handlers"
	"wedding_api/internal/router"
	"wedding_api/internal/service"
)

type fakeService struct {
	principalRole      string
	addProductReq      service.ProductRequest
	updateProductReq   service.UpdateProduct
	addPresenceReq     service.ConfirmationPresenceRequest
	updatePresenceReq  service.UpdateConfirmationPresence
	getProductID       int
	getPresenceID      int
	deleteProductID    int
	deletePresenceID   int
	confirmPresenceID  int
	cancelPresenceID   int
	productsResp       []*ent.Product
	productResp        *ent.Product
	productErr         error
	presencesResp      []*ent.ConfirmationPresence
	presenceResp       *ent.ConfirmationPresence
	presenceErr        error
	addProductResp     *ent.Product
	addProductErr      error
	updateProductResp  *ent.Product
	updateProductErr   error
	deleteProductErr   error
	addPresenceResp    *ent.ConfirmationPresence
	addPresenceErr     error
	updatePresenceResp *ent.ConfirmationPresence
	updatePresenceErr  error
	confirmResp        *ent.ConfirmationPresence
	confirmErr         error
	cancelResp         *ent.ConfirmationPresence
	cancelErr          error
	deletePresenceErr  error
}

func (f *fakeService) AddProduct(_ context.Context, product service.ProductRequest) (*ent.Product, error) {
	f.addProductReq = product
	return f.addProductResp, f.addProductErr
}

func (f *fakeService) ProductsList(_ context.Context) ([]*ent.Product, error) {
	return f.productsResp, nil
}

func (f *fakeService) GetProduct(_ context.Context, id int) (*ent.Product, error) {
	f.getProductID = id
	return f.productResp, f.productErr
}

func (f *fakeService) UpdateProduct(_ context.Context, newProduct service.UpdateProduct) (*ent.Product, error) {
	f.updateProductReq = newProduct
	return f.updateProductResp, f.updateProductErr
}

func (f *fakeService) ReserveProduct(_ context.Context, id int, reservedBy string) (*ent.Product, error) {
	f.updateProductReq = service.UpdateProduct{ID: id, ProductRequest: service.ProductRequest{ReservedBy: reservedBy}}
	return f.productResp, f.productErr
}

func (f *fakeService) DeleteProduct(_ context.Context, id int) error {
	f.deleteProductID = id
	return f.deleteProductErr
}

func (f *fakeService) AddConfirmationPresence(_ context.Context, presence service.ConfirmationPresenceRequest) (*ent.ConfirmationPresence, error) {
	f.addPresenceReq = presence
	return f.addPresenceResp, f.addPresenceErr
}

func (f *fakeService) ConfirmationPresencesList(_ context.Context) ([]*ent.ConfirmationPresence, error) {
	return f.presencesResp, nil
}

func (f *fakeService) GetConfirmationPresence(_ context.Context, id int) (*ent.ConfirmationPresence, error) {
	f.getPresenceID = id
	return f.presenceResp, f.presenceErr
}

func (f *fakeService) UpdateConfirmationPresence(_ context.Context, newPresence service.UpdateConfirmationPresence) (*ent.ConfirmationPresence, error) {
	f.updatePresenceReq = newPresence
	return f.updatePresenceResp, f.updatePresenceErr
}

func (f *fakeService) ConfirmPresence(_ context.Context, id int) (*ent.ConfirmationPresence, error) {
	f.confirmPresenceID = id
	return f.confirmResp, f.confirmErr
}

func (f *fakeService) CancelPresence(_ context.Context, id int) (*ent.ConfirmationPresence, error) {
	f.cancelPresenceID = id
	return f.cancelResp, f.cancelErr
}

func (f *fakeService) DeleteConfirmationPresence(_ context.Context, id int) error {
	f.deletePresenceID = id
	return f.deletePresenceErr
}

func (f *fakeService) AdminLogin(context.Context, service.AdminLoginRequest) (*service.LoginResult, error) {
	return nil, nil
}

func (f *fakeService) ExchangeFamilyToken(context.Context, string) (*service.LoginResult, error) {
	return nil, nil
}

func (f *fakeService) Logout(context.Context, string) error { return nil }

func (f *fakeService) CreateFamily(context.Context, service.FamilyRequest) (*ent.Family, error) {
	return nil, nil
}

func (f *fakeService) FamiliesList(context.Context) ([]*ent.Family, error) { return nil, nil }

func (f *fakeService) GetFamily(context.Context, int) (*ent.Family, error) { return nil, nil }

func (f *fakeService) GetMyFamily(context.Context) (*ent.Family, error) { return nil, nil }

func (f *fakeService) UpdateFamily(context.Context, int, service.FamilyRequest) (*ent.Family, error) {
	return nil, nil
}

func (f *fakeService) DeleteFamily(context.Context, int) error { return nil }

func (f *fakeService) CreateFamilyAccessLink(context.Context, int) (string, error) { return "", nil }

func (f *fakeService) RevokeFamilyAccessLink(context.Context, int) error { return nil }

func (f *fakeService) AuthenticateSession(context.Context, string) (*auth.Principal, error) {
	role := f.principalRole
	if role == "" {
		role = auth.RoleAdmin
	}
	return &auth.Principal{Role: role, SessionID: 1}, nil
}

func TestProtectedRoutesRequireSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	rec := httptest.NewRecorder()

	router.NewRouter(handlers.NewHandler(&fakeService{})).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestFamilyCannotUseAdminRoutes(t *testing.T) {
	fake := &fakeService{principalRole: auth.RoleFamily}
	req := httptest.NewRequest(http.MethodGet, "/api/admin/families", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "test-session"})
	rec := httptest.NewRecorder()

	router.NewRouter(handlers.NewHandler(fake)).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestAPIEndpoints(t *testing.T) {
	cases := []struct {
		name       string
		method     string
		path       string
		body       string
		setup      func(*fakeService)
		wantStatus int
		check      func(*testing.T, *fakeService, []byte)
	}{
		{
			name:   "list products",
			method: http.MethodGet,
			path:   "/api/products",
			setup: func(f *fakeService) {
				f.productsResp = []*ent.Product{{ID: 1, Title: "Mesa", Image: "img", Value: 10}}
			},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, _ *fakeService, body []byte) {
				var got []map[string]any
				mustDecode(t, body, &got)
				if len(got) != 1 || got[0]["title"] != "Mesa" {
					t.Fatalf("unexpected response: %s", string(body))
				}
			},
		},
		{
			name:   "legacy list products route",
			method: http.MethodGet,
			path:   "/products",
			setup: func(f *fakeService) {
				f.productsResp = []*ent.Product{{ID: 1, Title: "Mesa", Image: "img", Value: 10}}
			},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, _ *fakeService, body []byte) {
				var got []map[string]any
				mustDecode(t, body, &got)
				if len(got) != 1 || got[0]["title"] != "Mesa" {
					t.Fatalf("unexpected response: %s", string(body))
				}
			},
		},
		{
			name:   "get product",
			method: http.MethodGet,
			path:   "/api/products/1",
			setup: func(f *fakeService) {
				f.productResp = &ent.Product{ID: 1, Title: "Mesa", Image: "img", Value: 10}
			},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, f *fakeService, body []byte) {
				if f.getProductID != 1 {
					t.Fatalf("expected product id 1, got %d", f.getProductID)
				}
				var got map[string]any
				mustDecode(t, body, &got)
				if got["title"] != "Mesa" || got["value"] != float64(10) {
					t.Fatalf("unexpected response: %s", string(body))
				}
			},
		},
		{
			name:   "create product",
			method: http.MethodPost,
			path:   "/api/products",
			body:   `{"title":"Mesa","reserved_by":"Ana","image":"img","value":10}`,
			setup: func(f *fakeService) {
				f.addProductResp = &ent.Product{ID: 1, Title: "Mesa", Image: "img", Value: 10}
			},
			wantStatus: http.StatusCreated,
			check: func(t *testing.T, f *fakeService, body []byte) {
				if f.addProductReq.Title != "Mesa" || f.addProductReq.ReservedBy != "Ana" || f.addProductReq.Image != "img" || f.addProductReq.Value != 10 {
					t.Fatalf("request not forwarded correctly: %+v", f.addProductReq)
				}
				var got map[string]any
				mustDecode(t, body, &got)
				if got["title"] != "Mesa" {
					t.Fatalf("unexpected response: %s", string(body))
				}
			},
		},
		{
			name:   "update product",
			method: http.MethodPut,
			path:   "/api/products/1",
			body:   `{"title":"Mesa 2","reserved_by":"Ana","image":"img-2","value":20}`,
			setup: func(f *fakeService) {
				f.updateProductResp = &ent.Product{ID: 1, Title: "Mesa 2", Image: "img-2", Value: 20}
			},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, f *fakeService, body []byte) {
				if f.updateProductReq.ID != 1 || f.updateProductReq.Title != "Mesa 2" || f.updateProductReq.ReservedBy != "Ana" {
					t.Fatalf("request not forwarded correctly: %+v", f.updateProductReq)
				}
				var got map[string]any
				mustDecode(t, body, &got)
				if got["title"] != "Mesa 2" {
					t.Fatalf("unexpected response: %s", string(body))
				}
			},
		},
		{
			name:       "delete product",
			method:     http.MethodDelete,
			path:       "/api/products/1",
			setup:      func(f *fakeService) {},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, f *fakeService, _ []byte) {
				if f.deleteProductID != 1 {
					t.Fatalf("expected delete id 1, got %d", f.deleteProductID)
				}
			},
		},
		{
			name:   "get missing product",
			method: http.MethodGet,
			path:   "/api/products/9",
			setup: func(f *fakeService) {
				f.productErr = &ent.NotFoundError{}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "list presences",
			method: http.MethodGet,
			path:   "/api/presences",
			setup: func(f *fakeService) {
				f.presencesResp = []*ent.ConfirmationPresence{{ID: 1, Fullname: "Joao", PhotoBase64: "", IsConfirmed: false}}
			},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, _ *fakeService, body []byte) {
				var got []map[string]any
				mustDecode(t, body, &got)
				if len(got) != 1 || got[0]["fullname"] != "Joao" {
					t.Fatalf("unexpected response: %s", string(body))
				}
			},
		},
		{
			name:   "get presence",
			method: http.MethodGet,
			path:   "/api/presences/1",
			setup: func(f *fakeService) {
				f.presenceResp = &ent.ConfirmationPresence{ID: 1, Fullname: "Joao", PhotoBase64: "", IsConfirmed: false}
			},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, f *fakeService, body []byte) {
				if f.getPresenceID != 1 {
					t.Fatalf("expected presence id 1, got %d", f.getPresenceID)
				}
				var got map[string]any
				mustDecode(t, body, &got)
				if got["fullname"] != "Joao" {
					t.Fatalf("unexpected response: %s", string(body))
				}
			},
		},
		{
			name:   "create presence",
			method: http.MethodPost,
			path:   "/api/presences",
			body:   `{"fullname":"Joao da Silva","photo_base64":"data:image/png;base64,abc","is_confirmed":false}`,
			setup: func(f *fakeService) {
				f.addPresenceResp = &ent.ConfirmationPresence{ID: 1, Fullname: "Joao da Silva", PhotoBase64: "data:image/png;base64,abc", IsConfirmed: false}
			},
			wantStatus: http.StatusCreated,
			check: func(t *testing.T, f *fakeService, body []byte) {
				if f.addPresenceReq.Fullname != "Joao da Silva" || f.addPresenceReq.PhotoBase64 != "data:image/png;base64,abc" || f.addPresenceReq.IsConfirmed {
					t.Fatalf("request not forwarded correctly: %+v", f.addPresenceReq)
				}
				var got map[string]any
				mustDecode(t, body, &got)
				if got["fullname"] != "Joao da Silva" {
					t.Fatalf("unexpected response: %s", string(body))
				}
			},
		},
		{
			name:   "update presence",
			method: http.MethodPut,
			path:   "/api/presences/1",
			body:   `{"fullname":"Joao da Silva","photo_base64":"data:image/png;base64,xyz","is_confirmed":true}`,
			setup: func(f *fakeService) {
				f.updatePresenceResp = &ent.ConfirmationPresence{ID: 1, Fullname: "Joao da Silva", PhotoBase64: "data:image/png;base64,xyz", IsConfirmed: true}
			},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, f *fakeService, body []byte) {
				if f.updatePresenceReq.ID != 1 || f.updatePresenceReq.Fullname != "Joao da Silva" || !f.updatePresenceReq.IsConfirmed {
					t.Fatalf("request not forwarded correctly: %+v", f.updatePresenceReq)
				}
				var got map[string]any
				mustDecode(t, body, &got)
				if got["is_confirmed"] != true {
					t.Fatalf("unexpected response: %s", string(body))
				}
			},
		},
		{
			name:   "confirm presence",
			method: http.MethodPatch,
			path:   "/api/presences/1/confirm",
			setup: func(f *fakeService) {
				f.confirmResp = &ent.ConfirmationPresence{ID: 1, Fullname: "Joao", IsConfirmed: true}
			},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, f *fakeService, body []byte) {
				if f.confirmPresenceID != 1 {
					t.Fatalf("expected confirm id 1, got %d", f.confirmPresenceID)
				}
				var got map[string]any
				mustDecode(t, body, &got)
				if got["is_confirmed"] != true {
					t.Fatalf("unexpected response: %s", string(body))
				}
			},
		},
		{
			name:   "cancel presence",
			method: http.MethodPatch,
			path:   "/api/presences/1/cancel",
			setup: func(f *fakeService) {
				f.cancelResp = &ent.ConfirmationPresence{ID: 1, Fullname: "Joao", IsConfirmed: false}
			},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, f *fakeService, body []byte) {
				if f.cancelPresenceID != 1 {
					t.Fatalf("expected cancel id 1, got %d", f.cancelPresenceID)
				}
				var got map[string]any
				mustDecode(t, body, &got)
				if got["is_confirmed"] != false {
					t.Fatalf("unexpected response: %s", string(body))
				}
			},
		},
		{
			name:       "delete presence",
			method:     http.MethodDelete,
			path:       "/api/presences/1",
			setup:      func(f *fakeService) {},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, f *fakeService, _ []byte) {
				if f.deletePresenceID != 1 {
					t.Fatalf("expected delete id 1, got %d", f.deletePresenceID)
				}
			},
		},
		{
			name:   "get missing presence",
			method: http.MethodGet,
			path:   "/api/presences/9",
			setup: func(f *fakeService) {
				f.presenceErr = &ent.NotFoundError{}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeService{}
			if tc.setup != nil {
				tc.setup(fake)
			}

			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "test-session"})
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()

			router.NewRouter(handlers.NewHandler(fake)).ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d, body=%s", rec.Code, tc.wantStatus, rec.Body.String())
			}
			if tc.check != nil {
				tc.check(t, fake, rec.Body.Bytes())
			}
		})
	}
}

func TestCreatePresenceRejectsInvalidJSON(t *testing.T) {
	fake := &fakeService{}
	req := httptest.NewRequest(http.MethodPost, "/api/presences", strings.NewReader(`{"fullname":`))
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "test-session"})
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.NewRouter(handlers.NewHandler(fake)).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func mustDecode(t *testing.T, body []byte, out any) {
	t.Helper()
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(out); err != nil {
		t.Fatalf("failed to decode body %q: %v", string(body), err)
	}
}

func TestFakeServiceImplementsHandlerAPI(t *testing.T) {
	var _ handlers.ServiceAPI = (*fakeService)(nil)
}

func TestLegacyRoutesMirrorAPI(t *testing.T) {
	fake := &fakeService{productsResp: []*ent.Product{{ID: 1, Title: "Mesa", Image: "img", Value: 10}}}
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "test-session"})
	rec := httptest.NewRecorder()

	router.NewRouter(handlers.NewHandler(fake)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got []map[string]any
	mustDecode(t, rec.Body.Bytes(), &got)
	if got[0]["title"] != "Mesa" {
		t.Fatalf("unexpected response: %s", rec.Body.String())
	}
}
