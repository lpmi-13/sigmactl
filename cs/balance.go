package cs

import (
  "context"
  "github.com/lpmi-13/gocs"
)

// a wrapper for gocs.Balance
type Balance struct {
   *godo.Balance
}

// an interface to do things with the balance API endpoint(s)
type BalanceService interface {
  Get() (*Balance, error)
}

type balanceService struct {
   client *gocs.Client
}

var _ BalanceSErvice = &balanceService{}

// this builds a BalanceService instance
func NewBalanceService(gocsClient *gocs.Client) BalanceService {
   return &balanceService{
     client: gocsClient,
   }
}

func (as *balanceService) Get() (*Balance, error) {
    gocsBalance, _, err := as.client.Balance.Get(context.TODO())
   if err != nil {
      return nil, err
  }

   balance := &Balance{Balance: gocsBalance}
   return balance, nil
}
