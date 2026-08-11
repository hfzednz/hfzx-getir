package app
import ("context";"errors")
var ErrInvalid = errors.New("invalid argument")
type Deps struct{}
func (Deps) Pick(ctx context.Context, tenant, taskID string) (map[string]any, error) {
  if taskID=="" { return nil, ErrInvalid }
  return map[string]any{"taskId":taskID,"status":"picking"}, nil
}
func (Deps) Pack(ctx context.Context, tenant, taskID string) (map[string]any, error) {
  return map[string]any{"taskId":taskID,"status":"packed"}, nil
}
func (Deps) DispatchReady(ctx context.Context, tenant, taskID string) (map[string]any, error) {
  return map[string]any{"taskId":taskID,"status":"ready_for_dispatch"}, nil
}
