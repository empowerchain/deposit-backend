package scheme

import "context"

//encore:api auth method=GET
func AuthTest(_ context.Context) error {
	return nil
}

//encore:api public method=GET
func StatusTest(_ context.Context) error {
	return nil
}
