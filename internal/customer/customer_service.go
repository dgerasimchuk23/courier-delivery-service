package customer

import (
	"delivery/internal/models"
	"fmt"
)

type CustomerService struct {
	store *CustomerStore
}

func NewCustomerService(store *CustomerStore) *CustomerService {
	return &CustomerService{store: store}
}

func (s *CustomerService) Create(customer *models.Customer) error {
	if err := ValidateEmail(customer.Email); err != nil {
		return err
	}
	if err := ValidatePhone(customer.Phone); err != nil {
		return err
	}

	cust := models.Customer{
		Name:  customer.Name,
		Email: customer.Email,
		Phone: customer.Phone,
	}

	id, err := s.store.Add(cust)
	if err != nil {
		return err
	}
	customer.ID = id
	return nil
}

func (s *CustomerService) Get(id int) (*models.Customer, error) {
	customer, err := s.store.Get(id)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}
	return &models.Customer{
		ID:    customer.ID,
		Name:  customer.Name,
		Email: customer.Email,
		Phone: customer.Phone,
	}, nil
}

func (s *CustomerService) Update(id int, customer *models.Customer) error {
	cust := models.Customer{
		ID:    id,
		Name:  customer.Name,
		Email: customer.Email,
		Phone: customer.Phone,
	}

	if err := ValidateEmail(cust.Email); err != nil {
		return err
	}
	if err := ValidatePhone(cust.Phone); err != nil {
		return err
	}

	return s.store.Update(cust)
}

func (s *CustomerService) Delete(id int) error {
	return s.store.Delete(id)
}

func (s *CustomerService) List() ([]models.Customer, error) {
	customers, err := s.store.GetAll()
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении списка клиентов: %w", err)
	}

	var result []models.Customer
	for _, customer := range customers {
		result = append(result, models.Customer{
			ID:    customer.ID,
			Name:  customer.Name,
			Email: customer.Email,
			Phone: customer.Phone,
		})
	}
	return result, nil
}
