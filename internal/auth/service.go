package auth

import (
	"errors"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"nebula/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	ErrUserNotFound       = errors.New("用户不存在")
	ErrUserDisabled       = errors.New("用户已被禁用")
	ErrUserPending        = errors.New("用户待审核，请等待管理员批准")
	ErrInvalidToken       = errors.New("无效的 token")
	ErrTokenExpired       = errors.New("token 已过期")
	ErrUserAlreadyExists  = errors.New("用户名已存在")
	ErrEmailAlreadyExists = errors.New("邮箱已被注册")
	ErrWeakPassword       = errors.New("密码强度不足：必须包含字母和数字，长度 6-50 位")
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	RoleCode string `json:"role_code"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	User      UserInfo  `json:"user"`
}

type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	RoleCode string `json:"roleCode"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// validatePassword 校验密码强度
func validatePassword(password string) error {
	if len(password) < 6 {
		return errors.New("密码至少 6 位")
	}
	if len(password) > 50 {
		return errors.New("密码最多 50 位")
	}
	// 必须包含字母
	if !regexp.MustCompile(`[a-zA-Z]`).MatchString(password) {
		return errors.New("密码必须包含字母")
	}
	// 必须包含数字
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return errors.New("密码必须包含数字")
	}
	return nil
}

type Service struct {
	db          *gorm.DB
	jwtSecret   []byte
	tokenExpiry time.Duration
}

type Config struct {
	JWTSecret   string
	TokenExpiry time.Duration
}

func NewService(db *gorm.DB, cfg Config) *Service {
	if cfg.TokenExpiry == 0 {
		cfg.TokenExpiry = 24 * time.Hour
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "nebula-default-secret-key-change-in-production"
	}

	return &Service{
		db:          db,
		jwtSecret:   []byte(cfg.JWTSecret),
		tokenExpiry: cfg.TokenExpiry,
	}
}

func (s *Service) Login(req LoginRequest) (*LoginResponse, error) {
	var user models.User
	if err := s.db.Preload("Role").Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// 检查用户状态
	if user.Status == 0 {
		return nil, ErrUserPending
	}
	if user.Status != 1 {
		return nil, ErrUserDisabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	now := time.Now()
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role.Name,
		RoleCode: user.Role.Code,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	s.db.Model(&user).Update("last_login", now)

	return &LoginResponse{
		Token:     tokenString,
		ExpiresAt: now.Add(s.tokenExpiry),
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role.Name,
			RoleCode: user.Role.Code,
		},
	}, nil
}

func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.ExpiresAt.Before(time.Now()) {
			return nil, ErrTokenExpired
		}
		return claims, nil
	}

	return nil, ErrInvalidToken
}

func (s *Service) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// 校验新密码强度
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.db.Model(&user).Update("password", string(hashedPassword)).Error
}

func (s *Service) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Role").First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Role").Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) Register(req RegisterRequest) error {
	// 校验密码强度
	if err := validatePassword(req.Password); err != nil {
		return err
	}

	// 检查用户名是否存在
	var existingUser models.User
	if err := s.db.Where("username = ?", req.Username).First(&existingUser).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	} else {
		return ErrUserAlreadyExists
	}

	// 检查邮箱是否存在
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	} else {
		return ErrEmailAlreadyExists
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 创建用户（状态为待审核，不分配角色）
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Status:   0, // 0: 待审核，1: 正常，2: 禁用
		// RoleID 留空，待审核通过后再分配
	}

	return s.db.Create(&user).Error
}

// GetAllUsers 获取所有用户（管理员功能）
func (s *Service) GetAllUsers() ([]models.User, error) {
	var users []models.User
	if err := s.db.Preload("Role").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// SearchUsers 搜索用户
func (s *Service) SearchUsers(keyword string) ([]models.User, error) {
	var users []models.User
	query := s.db.Preload("Role")

	if keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUserStatus 更新用户状态（管理员功能）
func (s *Service) UpdateUserStatus(userID int, status int, operatorID uint, operator string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	oldStatus := user.Status

	// 更新用户状态
	if err := s.db.Model(&user).Update("status", status).Error; err != nil {
		return err
	}

	// 记录审核日志
	action := ""
	switch status {
	case 1:
		if oldStatus == 0 {
			action = "approve"
		} else {
			action = "enable"
		}
	case 2:
		action = "reject"
		if oldStatus == 1 {
			action = "disable"
		}
	}

	auditLog := models.AuditLog{
		UserID:     user.ID,
		Username:   user.Username,
		OperatorID: operatorID,
		Operator:   operator,
		Action:     action,
		OldStatus:  oldStatus,
		NewStatus:  status,
	}

	return s.db.Create(&auditLog).Error
}

// BatchUpdateUserStatus 批量更新用户状态
func (s *Service) BatchUpdateUserStatus(userIDs []int, status int, operatorID uint, operator string) error {
	// 批量更新
	if err := s.db.Model(&models.User{}).Where("id IN ?", userIDs).Update("status", status).Error; err != nil {
		return err
	}

	// 获取用户信息用于日志
	var users []models.User
	if err := s.db.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return err
	}

	// 批量记录日志
	action := ""
	switch status {
	case 1:
		action = "batch_approve"
	case 2:
		action = "batch_reject"
	}

	logs := make([]models.AuditLog, 0, len(users))
	for _, user := range users {
		logs = append(logs, models.AuditLog{
			UserID:     user.ID,
			Username:   user.Username,
			OperatorID: operatorID,
			Operator:   operator,
			Action:     action,
			OldStatus:  user.Status,
			NewStatus:  status,
		})
	}

	return s.db.Create(&logs).Error
}

// GetAuditLogs 获取审核日志
func (s *Service) GetAuditLogs(limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	if err := s.db.Order("created_at DESC").Limit(limit).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// GetAllRoles 获取所有角色
func (s *Service) GetAllRoles() ([]models.Role, error) {
	var roles []models.Role
	if err := s.db.Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// UpdateUserRole 更新用户角色
func (s *Service) UpdateUserRole(userID int, roleID int) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}
	return s.db.Model(&user).Update("role_id", roleID).Error
}

// UpdateUserProfile 更新用户资料
func (s *Service) UpdateUserProfile(userID int, email string, operatorID uint, operator string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	// 检查邮箱是否已被其他用户使用
	var existingUser models.User
	if err := s.db.Where("email = ? AND id != ?", email, userID).First(&existingUser).Error; err == nil {
		return errors.New("该邮箱已被使用")
	}

	// 更新用户邮箱
	if err := s.db.Model(&user).Update("email", email).Error; err != nil {
		return err
	}

	// 记录审核日志（如果是管理员修改）
	if operatorID != uint(userID) {
		_ = models.AuditLog{
			UserID:     user.ID,
			Username:   user.Username,
			OperatorID: operatorID,
			Operator:   operator,
			Action:     "update_profile",
			OldStatus:  user.Status,
			NewStatus:  user.Status,
			Remark:     "修改用户邮箱：" + email,
		}
	}

	return nil
}

// DeleteUser 删除用户
func (s *Service) DeleteUser(userID int, operatorID uint, operator string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	// 记录审核日志
	_ = models.AuditLog{
		UserID:     user.ID,
		Username:   user.Username,
		OperatorID: operatorID,
		Operator:   operator,
		Action:     "delete",
		OldStatus:  user.Status,
		NewStatus:  -1, // -1 表示删除
		Remark:     "用户已删除",
	}

	// 删除用户
	return s.db.Delete(&user).Error
}
