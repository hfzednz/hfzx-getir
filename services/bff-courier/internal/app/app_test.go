package app_test
import ("context";"testing";"github.com/nexora/bff-courier/internal/app")
func TestCourierJourney(t *testing.T) {
  d := app.Deps{}
  if _,err := d.Duty(context.Background(),"t","c1",true); err!=nil { t.Fatal(err) }
  if _,err := d.Offer(context.Background(),"t","c1","j1",true); err!=nil { t.Fatal(err) }
  if _,err := d.Complete(context.Background(),"t","j1"); err!=nil { t.Fatal(err) }
}
