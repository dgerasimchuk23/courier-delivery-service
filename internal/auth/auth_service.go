package auth

import (
	"context"
	"delivery/internal/business/models"
	"delivery/internal/cache"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Секретный ключ для подписи JWT токенов
const (
	jwtSecret       = "your-secret-key-here" // В реальном приложении должен быть в конфигурации
	accessTokenTTL  = 7 * time.Minute        // Время жизни access токена
	refreshTokenTTL = 30 * 24 * time.Hour    // Время жизни refresh токена (30 дней)
)

// Claims представляет данные, которые будут храниться в JWT токене
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// Предоставляет методы для аутентификации и авторизации
type AuthService struct {
	store         *UserStore
	cacheClient   *cache.RedisClient
	cleanupTicker *time.Ticker
	cleanupDone   chan bool
	mu            sync.Mutex
}

// Проверка, что AuthService реализует интерфейс AuthServiceInterface
var _ AuthServiceInterface = (*AuthService)(nil)

// Создание нового экземпляра AuthService
func NewAuthService(store *UserStore) *AuthService {
	service := &AuthService{
		store:       store,
		cleanupDone: make(chan bool),
	}

	// Запускаем периодическую очистку просроченных токенов
	service.startTokenCleanup()

	return service
}

// startTokenCleanup запускает периодическую очистку просроченных токенов
func (s *AuthService) startTokenCleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Останавливаем предыдущий тикер, если он был запущен
	if s.cleanupTicker != nil {
		s.cleanupTicker.Stop()
		s.cleanupDone <- true
	}

	// Создаем новый тикер
	s.cleanupTicker = time.NewTicker(12 * time.Hour)

	// Запускаем горутину для периодической очистки
	go func() {
		// Сразу выполняем очистку при запуске
		s.cleanupExpiredTokens()

		for {
			select {
			case <-s.cleanupTicker.C:
				s.cleanupExpiredTokens()
			case <-s.cleanupDone:
				return
			}
		}
	}()
}

// cleanupExpiredTokens удаляет просроченные токены
func (s *AuthService) cleanupExpiredTokens() {
	log.Println("Запуск очистки просроченных токенов...")
	if err := s.store.DeleteExpiredRefreshTokens(); err != nil {
		log.Printf("Ошибка при удалении просроченных токенов: %v", err)
	}
}

// WithCache добавляет клиент кэширования к сервису
func (s *AuthService) WithCache(cacheClient *cache.RedisClient) *AuthService {
	s.cacheClient = cacheClient
	return s
}

// Регистрация нового пользователя
func (s *AuthService) RegisterUser(email, password string) error {
	// Проверка валидности email и пароля
	if email == "" || password == "" {
		return errors.New("email и пароль не могут быть пустыми")
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("ошибка при хешировании пароля: %w", err)
	}

	// Создание пользователя
	user := models.User{
		Email:    email,
		Password: string(hashedPassword),
	}

	_, err = s.store.CreateUser(user)
	if err != nil {
		return fmt.Errorf("ошибка при регистрации пользователя: %w", err)
	}

	return nil
}

// Генерация токенов для пользователя
func (s *AuthService) GenerateTokens(userID int) (string, string, error) {
	// Генерация access токена
	accessTokenExpiry := time.Now().Add(accessTokenTTL)
	accessClaims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "delivery-app",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", fmt.Errorf("ошибка при создании access токена: %w", err)
	}

	// Генерация refresh токена с использованием UUID
	refreshToken := uuid.NewString()

	// Сохранение refresh токена в БД
	tokenExpiry := time.Now().UTC().Add(refreshTokenTTL)
	tokenModel := models.RefreshToken{
		UserID:    userID,
		Token:     refreshToken,
		ExpiresAt: tokenExpiry,
	}

	err = s.store.SaveRefreshToken(tokenModel)
	if err != nil {
		return "", "", fmt.Errorf("ошибка при сохранении refresh токена: %w", err)
	}

	return accessTokenString, refreshToken, nil
}

// Аутентификация пользователя и возвращение токенов
func (s *AuthService) LoginUser(email, password string) (string, string, error) {
	// Получение пользователя по email
	user, err := s.store.GetUserByEmail(email)
	if err != nil {
		return "", "", fmt.Errorf("ошибка при аутентификации: %w", err)
	}

	// Проверка пароля
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", errors.New("неверный пароль")
	}

	// Генерация токенов
	accessToken, refreshToken, err := s.GenerateTokens(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("ошибка при генерации токенов: %w", err)
	}

	return accessToken, refreshToken, nil
}

// Обновление токенов по refresh токену
func (s *AuthService) RefreshToken(refreshToken string) (string, string, error) {
	// Получаем refresh токен из БД
	token, err := s.store.GetRefreshToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("ошибка при получении refresh токена: %w", err)
	}

	// Удаляем старый refresh токен
	err = s.store.DeleteRefreshToken(refreshToken)
	if err != nil {
		log.Printf("Ошибка при удалении старого refresh токена: %v", err)
		// Продолжаем выполнение, так как это не критическая ошибка
	}

	// Генерируем новые токены
	accessToken, newRefreshToken, err := s.GenerateTokens(token.UserID)
	if err != nil {
		return "", "", fmt.Errorf("ошибка при генерации новых токенов: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

// Выход пользователя (логаут)
func (s *AuthService) Logout(accessToken, refreshToken string) error {
	// Удаляем refresh токен из БД
	err := s.store.DeleteRefreshToken(refreshToken)
	if err != nil {
		log.Printf("Ошибка при удалении refresh токена: %v", err)
		// Продолжаем выполнение, так как это не критическая ошибка
	}

	// Добавляем access токен в черный список, если доступен Redis
	if s.cacheClient != nil {
		ctx := context.Background()

		// Парсим токен для получения времени истечения
		token, err := jwt.ParseWithClaims(accessToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			// Проверяем метод подписи
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			// Если токен не может быть распарсен, просто добавляем его в черный список на стандартное время
			blacklistKey := fmt.Sprintf("blacklist:%s", accessToken)
			err := s.cacheClient.Set(ctx, blacklistKey, "revoked", accessTokenTTL)
			if err != nil {
				log.Printf("Ошибка при добавлении токена в черный список: %v", err)
				return nil
			}

			// Логируем блокировку токена
			log.Printf("Токен %s добавлен в черный список на %v", accessToken, accessTokenTTL)
			return nil
		}

		// Получаем время истечения токена
		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			// Вычисляем оставшееся время жизни токена
			expiresAt := claims.ExpiresAt.Time
			ttl := time.Until(expiresAt)
			if ttl < 0 {
				// Если токен уже истек, нет необходимости добавлять его в черный список
				return nil
			}

			// Добавляем токен в черный список на оставшееся время жизни
			blacklistKey := fmt.Sprintf("blacklist:%s", accessToken)
			err := s.cacheClient.Set(ctx, blacklistKey, "revoked", ttl)
			if err != nil {
				log.Printf("Ошибка при добавлении токена в черный список: %v", err)
				return nil
			}

			// Логируем блокировку токена
			log.Printf("Токен %s добавлен в черный список на %v", accessToken, ttl)
		}
	}

	return nil
}

// ValidateToken проверяет валидность токена и возвращает ID пользователя
func (s *AuthService) ValidateToken(tokenString string) (int, error) {
	// Проверяем, находится ли токен в черном списке
	if s.cacheClient != nil {
		ctx := context.Background()
		blacklistKey := fmt.Sprintf("blacklist:%s", tokenString)
		_, err := s.cacheClient.Get(ctx, blacklistKey)
		if err == nil {
			// Токен найден в черном списке
			return 0, errors.New("токен отозван")
		}
	}

	// Парсим токен
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return 0, fmt.Errorf("ошибка при проверке токена: %w", err)
	}

	// Проверяем валидность токена
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, errors.New("недействительный токен")
}

// Close закрывает ресурсы, используемые сервисом
func (s *AuthService) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cleanupTicker != nil {
		s.cleanupTicker.Stop()
		s.cleanupDone <- true
	}
}
