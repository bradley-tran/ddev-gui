//go:build linux
// +build linux

package main

/*
#cgo pkg-config: gtk+-3.0
#include <gtk/gtk.h>
#include <gdk/gdk.h>

extern void goMouseNav(int button);

// Custom GDK event handler that intercepts mouse back/forward buttons
// (GDK buttons 8 and 9) before WebKitGTK can consume them.
// All other events are passed through to GTK's default handler.
static void nav_event_handler(GdkEvent *event, gpointer data) {
    if (event->type == GDK_BUTTON_PRESS) {
        GdkEventButton *be = (GdkEventButton *)event;
        if (be->button == 8 || be->button == 9) {
            goMouseNav((int)be->button);
            return;
        }
    }
    // Consume the matching button release too, to avoid confusing GTK state
    if (event->type == GDK_BUTTON_RELEASE) {
        GdkEventButton *be = (GdkEventButton *)event;
        if (be->button == 8 || be->button == 9) {
            return;
        }
    }
    gtk_main_do_event(event);
}

static gboolean install_handler_idle(gpointer data) {
    gdk_event_handler_set(nav_event_handler, NULL, NULL);
    return G_SOURCE_REMOVE;
}

static void schedule_install_nav_handler() {
    g_idle_add(install_handler_idle, NULL);
}
*/
import "C"

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var navCtx context.Context

//export goMouseNav
func goMouseNav(button C.int) {
	if navCtx == nil {
		return
	}
	switch int(button) {
	case 8:
		go runtime.EventsEmit(navCtx, "mouse:back")
	case 9:
		go runtime.EventsEmit(navCtx, "mouse:forward")
	}
}

// InstallMouseNavHandler sets up a native GDK event filter to intercept mouse
// back/forward buttons (8/9) before WebKitGTK consumes them, and emits Wails
// events ("mouse:back" / "mouse:forward") to the frontend.
func InstallMouseNavHandler(ctx context.Context) {
	navCtx = ctx
	C.schedule_install_nav_handler()
}
