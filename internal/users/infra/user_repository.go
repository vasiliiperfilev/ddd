package infra

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/pkg/errors"
	"github.com/vasiliiperfilev/ddd/internal/common/auth"
	"github.com/vasiliiperfilev/ddd/internal/users/domain/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Model struct {
	Id          string
	Balance     int
	DisplayName string
	Role        auth.Role
	LastIP      string
}

type FirestoreUserRepository struct {
	firestoreClient *firestore.Client
}

func NewFirestoreUserRepository(firestoreClient *firestore.Client) *FirestoreUserRepository {
	if firestoreClient == nil {
		panic("missing firestoreClient")
	}

	return &FirestoreUserRepository{firestoreClient: firestoreClient}
}

func (f FirestoreUserRepository) Get(ctx context.Context, id string) (user.User, error) {
	model, err := f.getModel(
		func() (doc *firestore.DocumentSnapshot, err error) {
			return f.firestoreClient.Collection("users").Doc(id).Get(ctx)
		},
	)
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			return user.User{}, user.NotFoundError
		default:
			return user.User{}, err
		}
	}

	userModel, err := f.domainUserFromModel(model)
	if err != nil {
		return user.User{}, err
	}

	return userModel, nil
}

func (f FirestoreUserRepository) Update(ctx context.Context, id string, updateFn func(h user.User) (user.User, error)) error {
	err := f.firestoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		userDoc := f.firestoreClient.Collection("users").Doc(id)
		model, err := f.getModel(
			func() (doc *firestore.DocumentSnapshot, err error) {
				return tx.Get(userDoc)
			},
		)
		if err != nil {
			switch status.Code(err) {
			case codes.NotFound:
				return user.NotFoundError
			default:
				return err
			}
		}

		user, err := f.domainUserFromModel(model)
		if err != nil {
			return err
		}

		updatedUser, err := updateFn(user)
		if err != nil {
			return errors.Wrap(err, "unable to update hour")
		}

		return tx.Set(userDoc, updatedUser)
	})

	return errors.Wrap(err, "firestore transaction failed")
}

func (f FirestoreUserRepository) getModel(
	getDocumentFn func() (doc *firestore.DocumentSnapshot, err error),
) (Model, error) {
	doc, err := getDocumentFn()
	if err != nil {
		return Model{}, err
	}

	var userModel Model
	err = doc.DataTo(&userModel)
	if err != nil {
		return Model{}, err
	}

	return userModel, nil
}

func (f FirestoreUserRepository) domainUserFromModel(userModel Model) (user.User, error) {
	user := user.User{
		Id:          userModel.Id,
		Balance:     userModel.Balance,
		DisplayName: userModel.DisplayName,
		Role:        userModel.Role,
		LastIP:      userModel.LastIP,
	}
	return user, nil
}
