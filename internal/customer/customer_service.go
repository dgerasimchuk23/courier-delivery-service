package customer

type CustomerService struct {
	store *CustomerStore
}

func NewCustomerService(store *CustomerStore) *CustomerService {
	return &CustomerService{store: store}
}

func (s *CustomerService) RegisterCustomer(name, email, phone string) (Customer, error) {
	// Валидация данных
	if err := ValidateEmail(email); err != nil {
		return Customer{}, err
	}
	if err := ValidatePhone(phone); err != nil {
		return Customer{}, err
	}

	customer := Customer{
		Name:  name,
		Email: email,
		Phone: phone,
	}

	// Добавление клиента в хранилище данных
	id, err := s.store.Add(customer)
	if err != nil {
		return Customer{}, err
	}
	customer.ID = id

	return customer, nil
}

func (s *CustomerService) GetCustomer(id int) (Customer, error) {
	return s.store.Get(id)
}

func (s *CustomerService) UpdateCustomer(id int, name, email, phone string) error {
	// Валидация данных
	if err := ValidateEmail(email); err != nil {
		return err
	}
	if err := ValidatePhone(phone); err != nil {
		return err
	}

	// Обновление данных клиента
	customer := Customer{
		ID:    id,
		Name:  name,
		Email: email,
		Phone: phone,
	}
	return s.store.Update(customer)
}

func (s *CustomerService) DeleteCustomer(id int) error {
	return s.store.Delete(id)
}

func (s *CustomerService) String() string {
	return "CustomerService {работает с клиентами}"
}
