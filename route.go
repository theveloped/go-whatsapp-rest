package main

import (
	ctl "github.com/theveloped/go-whatsapp-rest/controller"
	svc "github.com/theveloped/go-whatsapp-rest/service"
	"github.com/go-chi/chi"
)

// RoutesInit Function
func routesInit() {
	// Set Endpoint for Root Functions
	svc.Router.Get(svc.RouterBasePath, ctl.GetIndex)
	svc.Router.Get(svc.RouterBasePath+"/health", ctl.GetHealth)

	// Set Endpoint for Authorization Functions
	svc.Router.With(svc.AuthBasic).Get(svc.RouterBasePath+"/auth", ctl.GetAuth)

	// Set Endpoint for WhatsApp Functions
	svc.Router.With(svc.AuthJWT).Post(svc.RouterBasePath+"/login", ctl.WhatsAppLogin)
	svc.Router.With(svc.AuthJWT).Post(svc.RouterBasePath+"/messagetext", ctl.WhatsAppSendText)
	svc.Router.With(svc.AuthJWT).Post(svc.RouterBasePath+"/messageimage", ctl.WhatsAppSendImage)
	svc.Router.With(svc.AuthJWT).Post(svc.RouterBasePath+"/logout", ctl.WhatsAppLogout)

	// Restful endpoints
	svc.Router.Route(svc.RouterBasePath + "/messages", func(r chi.Router) {
		r.With(svc.AuthJWT).Get("/{messageID}/data", ctl.WhatsAppGetAttachment)
		r.With(svc.AuthJWT).Post("/", ctl.WhatsAppSendGeneric)
	})
}
