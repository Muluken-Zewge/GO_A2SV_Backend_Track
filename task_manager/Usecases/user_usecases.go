package usecases

import (
	"context"
	"errors"
	"fmt"
	domain "taskmanager/Domain"
	infrastructure "taskmanager/Infrastructure"
	repositories "taskmanager/Repositories"

	"github.com/google/uuid"
)

type UserUsecase interface {
	RegisterUser(ctx context.Context, userName string, password string) (domain.User, error)
	AuthenticateUser(ctx context.Context, userName string, password string) (string, error)
	PromoteUser(ctx context.Context, userId string) (domain.User, error)
}

type UserUsecaseImpl struct {
	userRepository repositories.UserRepository
}

// Constructor for dependency injection
func NewUserUsecase(repo repositories.UserRepository) UserUsecase {
	return &UserUsecaseImpl{
		userRepository: repo,
	}
}

func (u *UserUsecaseImpl) RegisterUser(ctx context.Context, userName string, password string) (domain.User, error) {

	// validate password length
	if len(password) < 4 {
		return domain.User{}, fmt.Errorf("%w:password should be at least 4 characters", domain.ErrValidation)
	}

	// check if user name is available
	err := u.userRepository.IsUsernameAvailable(ctx, userName)
	if err != nil {
		if errors.Is(err, domain.ErrAleadyExists) {
			return domain.User{}, fmt.Errorf("%w:username already exists", domain.ErrAleadyExists)
		}
		return domain.User{}, err
	}

	// hash password
	hashedPassword, err := infrastructure.HashPassword(password)
	if err != nil {
		return domain.User{}, err
	}

	// check if database empty
	isEmpty, err := u.userRepository.IsDatabaseEmpty(ctx)
	if err != nil {
		return domain.User{}, err
	}

	// create new id(uuid)
	newId := uuid.New()

	// create a user variable and assign all the values
	var newUser domain.User
	newUser.ID = newId
	newUser.UserName = userName
	newUser.HashedPassword = hashedPassword
	if isEmpty {
		newUser.Role = domain.RoleAdmin
	} else {
		newUser.Role = domain.RoleUser
	}

	// save to database
	savedUser, err := u.userRepository.SaveUser(ctx, newUser)
	if err != nil {
		return domain.User{}, err
	}

	return savedUser, nil
}

func (u *UserUsecaseImpl) AuthenticateUser(ctx context.Context, userName string, password string) (string, error) {

	// check if username exists
	userId, SavedPassword, role, err := u.userRepository.DoesUserExist(ctx, userName)
	if err != nil {
		return "", err
	}

	// check if password is correct
	err = infrastructure.ComparePassword(SavedPassword, password)
	if err != nil {
		return "", err
	}

	// generate jwt token
	token, err := infrastructure.GenerateJWT(userId, userName, role)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (u *UserUsecaseImpl) PromoteUser(ctx context.Context, userId string) (domain.User, error) {

	// call the promote user function from repository
	promotedUser, err := u.userRepository.PromoteUser(ctx, userId)
	if err != nil {
		return domain.User{}, err
	}

	return promotedUser, nil
}
