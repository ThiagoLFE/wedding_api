package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"wedding_api/ent"
	"wedding_api/ent/confirmationpresence"
	"wedding_api/ent/family"
	"wedding_api/ent/familyaccesstoken"
	"wedding_api/ent/predicate"
	"wedding_api/ent/product"
	"wedding_api/ent/session"
	"wedding_api/ent/user"
	"wedding_api/internal/auth"
)

const sessionDuration = 30 * 24 * time.Hour

type Service struct {
	client *ent.Client
}

func NewService(client *ent.Client) *Service {
	return &Service{client: client}
}

type ProductRequest struct {
	Title      string  `json:"title"`
	ReservedBy string  `json:"reserved_by"`
	Image      string  `json:"image"`
	Value      float64 `json:"value"`
}

type UpdateProduct struct {
	ID int `json:"id"`
	ProductRequest
}

func (p *ProductRequest) Validate() error {
	if len(strings.TrimSpace(p.Title)) == 0 {
		return errors.New("the Product must have a title")
	}
	if p.Value <= 0.00 {
		return errors.New("product value must be greater than 0")
	}
	return nil
}

type ConfirmationPresenceRequest struct {
	FamilyID    int    `json:"family_id"`
	Fullname    string `json:"fullname"`
	PhotoBase64 string `json:"photo_base64"`
	IsConfirmed bool   `json:"is_confirmed"`
}

type UpdateConfirmationPresence struct {
	ID int `json:"id"`
	ConfirmationPresenceRequest
}

func (p *ConfirmationPresenceRequest) Validate() error {
	if len(strings.TrimSpace(p.Fullname)) == 0 {
		return errors.New("fullname is required")
	}
	return nil
}

type FamilyRequest struct {
	Name string `json:"name"`
}

func (f *FamilyRequest) Validate() error {
	if strings.TrimSpace(f.Name) == "" {
		return errors.New("family name is required")
	}
	return nil
}

type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResult struct {
	Token     string
	ExpiresAt time.Time
	Principal *auth.Principal
}

func (s *Service) EnsureAdmin(ctx context.Context, email, password string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return nil
	}
	count, err := s.client.User.Query().Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	_, err = s.client.User.Create().
		SetEmail(email).
		SetPasswordHash(hash).
		SetRole(user.RoleAdmin).
		Save(ctx)
	return err
}

func (s *Service) AdminLogin(ctx context.Context, request AdminLoginRequest) (*LoginResult, error) {
	email := strings.ToLower(strings.TrimSpace(request.Email))
	admin, err := s.client.User.Query().Where(user.EmailEQ(email)).Only(ctx)
	if err != nil || admin.Role != user.RoleAdmin || auth.CheckPassword(admin.PasswordHash, request.Password) != nil {
		return nil, errors.New("invalid credentials")
	}

	return s.createSession(ctx, auth.RoleAdmin, &admin.ID, nil)
}

func (s *Service) AuthenticateSession(ctx context.Context, rawToken string) (*auth.Principal, error) {
	if strings.TrimSpace(rawToken) == "" {
		return nil, auth.ErrInvalidToken
	}

	stored, err := s.client.Session.Query().
		Where(session.TokenHashEQ(auth.HashToken(rawToken))).
		WithUser().
		WithFamily().
		Only(ctx)
	if err != nil || stored.RevokedAt != nil || !stored.ExpiresAt.After(time.Now()) {
		return nil, auth.ErrInvalidToken
	}

	principal := &auth.Principal{SessionID: stored.ID, Role: stored.Role.String()}
	if stored.Edges.User != nil {
		principal.UserID = &stored.Edges.User.ID
	}
	if stored.Edges.Family != nil {
		principal.FamilyID = &stored.Edges.Family.ID
	}
	return principal, nil
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	now := time.Now()
	_, err := s.client.Session.Update().
		Where(session.TokenHashEQ(auth.HashToken(rawToken)), session.RevokedAtIsNil()).
		SetRevokedAt(now).
		Save(ctx)
	return err
}

func (s *Service) ExchangeFamilyToken(ctx context.Context, rawToken string) (*LoginResult, error) {
	stored, err := s.client.FamilyAccessToken.Query().
		Where(
			familyaccesstoken.TokenHashEQ(auth.HashToken(rawToken)),
			familyaccesstoken.RevokedAtIsNil(),
		).
		WithFamily().
		Only(ctx)
	if err != nil || stored.Edges.Family == nil || (stored.ExpiresAt != nil && !stored.ExpiresAt.After(time.Now())) {
		return nil, auth.ErrInvalidToken
	}

	now := time.Now()
	if err := stored.Update().SetLastUsedAt(now).Exec(ctx); err != nil {
		return nil, err
	}
	return s.createSession(ctx, auth.RoleFamily, nil, &stored.Edges.Family.ID)
}

func (s *Service) createSession(ctx context.Context, role string, userID, familyID *int) (*LoginResult, error) {
	rawToken, err := auth.GenerateToken()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(sessionDuration)
	builder := s.client.Session.Create().
		SetTokenHash(auth.HashToken(rawToken)).
		SetRole(session.Role(role)).
		SetExpiresAt(expiresAt)
	if userID != nil {
		builder.SetUserID(*userID)
	}
	if familyID != nil {
		builder.SetFamilyID(*familyID)
	}
	created, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return &LoginResult{
		Token:     rawToken,
		ExpiresAt: expiresAt,
		Principal: &auth.Principal{SessionID: created.ID, UserID: userID, FamilyID: familyID, Role: role},
	}, nil
}

// Families
func (s *Service) CreateFamily(ctx context.Context, request FamilyRequest) (*ent.Family, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := request.Validate(); err != nil {
		return nil, err
	}
	return s.client.Family.Create().SetName(strings.TrimSpace(request.Name)).Save(ctx)
}

func (s *Service) FamiliesList(ctx context.Context) ([]*ent.Family, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	return s.client.Family.Query().WithPresences(func(query *ent.ConfirmationPresenceQuery) {
		query.WithFamily()
	}).Order(ent.Asc(family.FieldID)).All(ctx)
}

func (s *Service) GetFamily(ctx context.Context, id int) (*ent.Family, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	return s.client.Family.Query().Where(family.IDEQ(id)).WithPresences(func(query *ent.ConfirmationPresenceQuery) {
		query.WithFamily()
	}).Only(ctx)
}

func (s *Service) GetMyFamily(ctx context.Context) (*ent.Family, error) {
	principal, err := auth.RequireRole(ctx, auth.RoleFamily)
	if err != nil || principal.FamilyID == nil {
		return nil, auth.ErrForbidden
	}
	return s.GetFamily(ctx, *principal.FamilyID)
}

func (s *Service) UpdateFamily(ctx context.Context, id int, request FamilyRequest) (*ent.Family, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := request.Validate(); err != nil {
		return nil, err
	}
	return s.client.Family.UpdateOneID(id).SetName(strings.TrimSpace(request.Name)).Save(ctx)
}

func (s *Service) DeleteFamily(ctx context.Context, id int) error {
	if err := requireAdmin(ctx); err != nil {
		return err
	}
	if _, err := s.client.Family.Query().Where(family.IDEQ(id)).Only(ctx); err != nil {
		return err
	}
	if _, err := s.client.ConfirmationPresence.Delete().Where(confirmationpresence.HasFamilyWith(family.IDEQ(id))).Exec(ctx); err != nil {
		return err
	}
	if _, err := s.client.FamilyAccessToken.Delete().Where(familyaccesstoken.HasFamilyWith(family.IDEQ(id))).Exec(ctx); err != nil {
		return err
	}
	if _, err := s.client.Session.Delete().Where(session.HasFamilyWith(family.IDEQ(id))).Exec(ctx); err != nil {
		return err
	}
	return s.client.Family.DeleteOneID(id).Exec(ctx)
}

func (s *Service) CreateFamilyAccessLink(ctx context.Context, familyID int) (string, error) {
	if _, err := auth.RequireRole(ctx, auth.RoleAdmin); err != nil {
		return "", err
	}
	if _, err := s.client.Family.Query().Where(family.IDEQ(familyID)).Only(ctx); err != nil {
		return "", err
	}
	now := time.Now()
	if _, err := s.client.FamilyAccessToken.Update().
		Where(familyaccesstoken.HasFamilyWith(family.IDEQ(familyID)), familyaccesstoken.RevokedAtIsNil()).
		SetRevokedAt(now).
		Save(ctx); err != nil {
		return "", err
	}
	if _, err := s.client.Session.Update().
		Where(session.HasFamilyWith(family.IDEQ(familyID)), session.RevokedAtIsNil()).
		SetRevokedAt(now).
		Save(ctx); err != nil {
		return "", err
	}
	rawToken, err := auth.GenerateToken()
	if err != nil {
		return "", err
	}
	_, err = s.client.FamilyAccessToken.Create().
		SetTokenHash(auth.HashToken(rawToken)).
		SetFamilyID(familyID).
		Save(ctx)
	return rawToken, err
}

func (s *Service) RevokeFamilyAccessLink(ctx context.Context, familyID int) error {
	if _, err := auth.RequireRole(ctx, auth.RoleAdmin); err != nil {
		return err
	}
	if _, err := s.client.Family.Query().Where(family.IDEQ(familyID)).Only(ctx); err != nil {
		return err
	}
	now := time.Now()
	if _, err := s.client.FamilyAccessToken.Update().
		Where(familyaccesstoken.HasFamilyWith(family.IDEQ(familyID)), familyaccesstoken.RevokedAtIsNil()).
		SetRevokedAt(now).
		Save(ctx); err != nil {
		return err
	}
	// Revoking the link also cuts off sessions created by that family link.
	_, err := s.client.Session.Update().
		Where(session.HasFamilyWith(family.IDEQ(familyID)), session.RevokedAtIsNil()).
		SetRevokedAt(now).
		Save(ctx)
	return err
}

// Products are shared by all families. Only admins can manage them; families
// can reserve an existing product through the dedicated operation below.
func (s *Service) AddProduct(ctx context.Context, request ProductRequest) (*ent.Product, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := request.Validate(); err != nil {
		return nil, err
	}
	return s.client.Product.Create().
		SetTitle(strings.TrimSpace(request.Title)).
		SetValue(request.Value).
		SetReservedBy(strings.TrimSpace(request.ReservedBy)).
		SetImage(request.Image).
		Save(ctx)
}

func (s *Service) ProductsList(ctx context.Context) ([]*ent.Product, error) {
	return s.client.Product.Query().Order(ent.Asc(product.FieldID)).All(ctx)
}

func (s *Service) GetProduct(ctx context.Context, id int) (*ent.Product, error) {
	return s.client.Product.Query().Where(product.IDEQ(id)).Only(ctx)
}

func (s *Service) UpdateProduct(ctx context.Context, newProduct UpdateProduct) (*ent.Product, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := newProduct.ProductRequest.Validate(); err != nil {
		return nil, err
	}
	return s.client.Product.UpdateOneID(newProduct.ID).
		SetTitle(strings.TrimSpace(newProduct.Title)).
		SetValue(newProduct.Value).
		SetReservedBy(strings.TrimSpace(newProduct.ReservedBy)).
		SetImage(newProduct.Image).
		Save(ctx)
}

func (s *Service) ReserveProduct(ctx context.Context, id int, reservedBy string) (*ent.Product, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.Role != auth.RoleFamily && principal.Role != auth.RoleAdmin {
		return nil, auth.ErrForbidden
	}
	reservedBy = strings.TrimSpace(reservedBy)
	if reservedBy == "" {
		return nil, errors.New("reserved_by is required")
	}
	if principal.Role == auth.RoleAdmin {
		return s.client.Product.UpdateOneID(id).SetReservedBy(reservedBy).Save(ctx)
	}
	updated, err := s.client.Product.Update().
		Where(
			product.IDEQ(id),
			product.Or(product.ReservedByEQ(""), predicate.Product(sql.FieldIsNull(product.FieldReservedBy))),
		).
		SetReservedBy(reservedBy).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	if updated == 0 {
		if _, err := s.client.Product.Query().Where(product.IDEQ(id)).Only(ctx); err != nil {
			return nil, err
		}
		return nil, errors.New("product is already reserved")
	}
	return s.client.Product.Query().Where(product.IDEQ(id)).Only(ctx)
}

func (s *Service) DeleteProduct(ctx context.Context, id int) error {
	if err := requireAdmin(ctx); err != nil {
		return err
	}
	return s.client.Product.DeleteOneID(id).Exec(ctx)
}

// Confirmation presences
func (s *Service) AddConfirmationPresence(ctx context.Context, request ConfirmationPresenceRequest) (*ent.ConfirmationPresence, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := request.Validate(); err != nil {
		return nil, err
	}
	if request.FamilyID <= 0 {
		return nil, errors.New("family_id is required")
	}
	return s.client.ConfirmationPresence.Create().
		SetFamilyID(request.FamilyID).
		SetFullname(strings.TrimSpace(request.Fullname)).
		SetPhotoBase64(request.PhotoBase64).
		SetIsConfirmed(request.IsConfirmed).
		Save(ctx)
}

func (s *Service) ConfirmationPresencesList(ctx context.Context) ([]*ent.ConfirmationPresence, error) {
	query := s.client.ConfirmationPresence.Query().WithFamily().Order(ent.Asc(confirmationpresence.FieldID))
	if principal, ok := auth.PrincipalFromContext(ctx); ok && principal.Role == auth.RoleFamily && principal.FamilyID != nil {
		query = query.Where(confirmationpresence.HasFamilyWith(family.IDEQ(*principal.FamilyID)))
	}
	return query.All(ctx)
}

func (s *Service) GetConfirmationPresence(ctx context.Context, id int) (*ent.ConfirmationPresence, error) {
	query := s.client.ConfirmationPresence.Query().Where(confirmationpresence.IDEQ(id)).WithFamily()
	if principal, ok := auth.PrincipalFromContext(ctx); ok && principal.Role == auth.RoleFamily && principal.FamilyID != nil {
		query = query.Where(confirmationpresence.HasFamilyWith(family.IDEQ(*principal.FamilyID)))
	}
	return query.Only(ctx)
}

func (s *Service) UpdateConfirmationPresence(ctx context.Context, newPresence UpdateConfirmationPresence) (*ent.ConfirmationPresence, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := newPresence.ConfirmationPresenceRequest.Validate(); err != nil {
		return nil, err
	}
	if newPresence.FamilyID <= 0 {
		return nil, errors.New("family_id is required")
	}
	return s.client.ConfirmationPresence.UpdateOneID(newPresence.ID).
		SetFamilyID(newPresence.FamilyID).
		SetFullname(strings.TrimSpace(newPresence.Fullname)).
		SetPhotoBase64(newPresence.PhotoBase64).
		SetIsConfirmed(newPresence.IsConfirmed).
		Save(ctx)
}

func (s *Service) ConfirmPresence(ctx context.Context, id int) (*ent.ConfirmationPresence, error) {
	return s.updatePresenceConfirmation(ctx, id, true)
}

func (s *Service) CancelPresence(ctx context.Context, id int) (*ent.ConfirmationPresence, error) {
	return s.updatePresenceConfirmation(ctx, id, false)
}

func (s *Service) updatePresenceConfirmation(ctx context.Context, id int, confirmed bool) (*ent.ConfirmationPresence, error) {
	query := s.client.ConfirmationPresence.Query().Where(confirmationpresence.IDEQ(id))
	if principal, ok := auth.PrincipalFromContext(ctx); ok && principal.Role == auth.RoleFamily && principal.FamilyID != nil {
		query = query.Where(confirmationpresence.HasFamilyWith(family.IDEQ(*principal.FamilyID)))
	}
	presence, err := query.Only(ctx)
	if err != nil {
		return nil, err
	}
	return presence.Update().SetIsConfirmed(confirmed).Save(ctx)
}

func (s *Service) DeleteConfirmationPresence(ctx context.Context, id int) error {
	if err := requireAdmin(ctx); err != nil {
		return err
	}
	return s.client.ConfirmationPresence.DeleteOneID(id).Exec(ctx)
}

func requireAdmin(ctx context.Context) error {
	_, err := auth.RequireRole(ctx, auth.RoleAdmin)
	return err
}

// Compile-time checks for the router/handler contracts.
var _ auth.SessionAuthenticator = (*Service)(nil)
