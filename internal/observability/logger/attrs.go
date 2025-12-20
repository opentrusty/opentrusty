package logger

import "log/slog"

// Common attribute keys for consistent logging across the application

// Request attributes
func RequestID(id string) slog.Attr {
	return slog.String("request_id", id)
}

func Method(method string) slog.Attr {
	return slog.String("method", method)
}

func Path(path string) slog.Attr {
	return slog.String("path", path)
}

func RemoteAddr(addr string) slog.Attr {
	return slog.String("remote_addr", addr)
}

func UserAgent(ua string) slog.Attr {
	return slog.String("user_agent", ua)
}

func StatusCode(code int) slog.Attr {
	return slog.Int("status_code", code)
}

func Duration(ms int64) slog.Attr {
	return slog.Int64("duration_ms", ms)
}

// Identity attributes
func UserID(id string) slog.Attr {
	return slog.String("user_id", id)
}

func Email(email string) slog.Attr {
	return slog.String("email", email)
}

func SessionID(id string) slog.Attr {
	return slog.String("session_id", id)
}

// OAuth/OIDC attributes
func ClientID(id string) slog.Attr {
	return slog.String("client_id", id)
}

func Scope(scope string) slog.Attr {
	return slog.String("scope", scope)
}

func GrantType(grantType string) slog.Attr {
	return slog.String("grant_type", grantType)
}

func RedirectURI(uri string) slog.Attr {
	return slog.String("redirect_uri", uri)
}

// Error attributes
func Error(err error) slog.Attr {
	if err == nil {
		return slog.String("error", "")
	}
	return slog.String("error", err.Error())
}

func ErrorType(errType string) slog.Attr {
	return slog.String("error_type", errType)
}

// Database attributes
func Query(query string) slog.Attr {
	return slog.String("query", query)
}

func RowsAffected(rows int64) slog.Attr {
	return slog.Int64("rows_affected", rows)
}

// Component attributes
func Component(name string) slog.Attr {
	return slog.String("component", name)
}

func Operation(op string) slog.Attr {
	return slog.String("operation", op)
}

// String creates a generic string attribute
func String(key, value string) slog.Attr {
	return slog.String(key, value)
}
