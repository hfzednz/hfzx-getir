package app_test
import ("context";"testing";"github.com/nexora/bff-warehouse/internal/app")
func TestWarehouseJourney(t *testing.T) {
  d:=app.Deps{}
  if _,err:=d.Pick(context.Background(),"t","task1"); err!=nil { t.Fatal(err) }
  if _,err:=d.Pack(context.Background(),"t","task1"); err!=nil { t.Fatal(err) }
  if _,err:=d.DispatchReady(context.Background(),"t","task1"); err!=nil { t.Fatal(err) }
}
