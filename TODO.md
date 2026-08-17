* migrate code as HA addon
* rework widgets as isolated component
* custom themes
* json not passed directly to backend
* make sure mobile view works
* classify main sections in the admin SPA (e.g. notes/images = user content, backgrounds/themes = theming)
* uploaded themes only become usable by widgets/renderers after a backend restart (icons/fonts registered once at startup in app/router/main.go)
