package displayers

import (
  "io"
  "time"

  "github.com/lpmi-13/sigmactl/cs"
)

type Balance struct {
    *cs.Balance
}


var _ Displayable = &Balance{}

func (a *Balance) JSON(out io.Writer) error {
  return writeJSON(a.Balance, out)
}

// these are the names of the columns in the output
func (a *Balance) Cols() []string {
   return []string{
      "TotalBalance", "Currency"
   }
}

func (a *Balance) ColMap() map[string]string {
  return map[string]string{
     "TotalBalance": "Total Balance",
     "Currency": "Currency",
   }
}

func (a *Balance) KV() []map[string]interface{} {
   out := []map[string]interface{}{}
    x := map[string]interface{}{
          "TotalBalance": a.TotalBalance,
          "Currency": a.Currency,
    }
    out = append(out, x)

    return out
}
