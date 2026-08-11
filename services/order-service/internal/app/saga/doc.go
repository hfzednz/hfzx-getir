// Package saga documents the place-order saga runner.
//
// The PlaceOrder saga runner lives in package app (place.go) so it can own
// methods on *app.Deps. Steps (with retry stub):
//
//	Validate → SoftReserve → AuthorizePayment → ConfirmHard → StartFulfillment
//
// Later warehouse/dispatch events advance state beyond warehouse_assigned.
package saga
