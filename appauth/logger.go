package appauth

func (a *Auth) logf(format string, v ...any) {
	if a.cfg.Logger != nil {
		a.cfg.Logger.Printf(format, v...)
	}
}
