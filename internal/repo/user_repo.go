package repo

import (
	"context"

	db "IFMS-be/sql/generated"

	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepo struct {
	q *db.Queries
}

func NewUserRepo(q *db.Queries) *UserRepo {
	return &UserRepo{q: q}
}

func (r *UserRepo) GetByID(ctx context.Context, id pgtype.UUID) (db.User, error) {
	return r.q.GetUserByID(ctx, id)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (db.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (db.User, error) {
	return r.q.GetUserByUsername(ctx, username)
}

func (r *UserRepo) List(ctx context.Context, limit, offset int32) ([]db.User, error) {
	return r.q.ListUsers(ctx, db.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *UserRepo) Create(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	return r.q.CreateUser(ctx, arg)
}

func (r *UserRepo) Update(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	return r.q.UpdateUser(ctx, arg)
}

func (r *UserRepo) Delete(ctx context.Context, id pgtype.UUID) error {
	return r.q.DeleteUser(ctx, id)
}
