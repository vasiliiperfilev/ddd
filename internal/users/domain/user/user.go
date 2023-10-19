package user

import (
	"errors"

	"github.com/vasiliiperfilev/ddd/internal/common/auth"
)

type User struct {
	Id          string
	Balance     int
	DisplayName string
	Role        auth.Role
	LastIP      string
}

type ResponseDto struct {
	Id          string    `json:"id"`
	Balance     int       `json:"balance"`
	DisplayName string    `json:"displayName"`
	Role        auth.Role `json:"role"`
}

func (u *User) toResponseDto() ResponseDto {
	return ResponseDto{
		Id:          u.Id,
		Balance:     u.Balance,
		DisplayName: u.DisplayName,
		Role:        u.Role,
	}
}

func (u *User) UpdateLastIP(ip string) {
	u.LastIP = ip
}

func (u *User) UpdateBalance(amountChange int) error {
	u.Balance += amountChange
	if u.Balance < 0 {
		return errors.New("balance cannot be smaller than 0")
	}
	return nil
}
