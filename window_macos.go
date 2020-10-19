// +build darwin
// +build amd64,!sim

package wuapp

/*
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework Cocoa -framework WebKit

#include <stdlib.h>
#include <string.h>
#include <Cocoa/Cocoa.h>
#include <mach-o/dyld.h>
#include <WebKit/WebKit.h>
#include <objc/runtime.h>
#include "util_cocoa.h"
#include "menu_macos.h"
#include "window.h"

extern void receive(const char* s);

@interface MessageHandler : NSObject <WKScriptMessageHandler> {
}
@end

@implementation MessageHandler
- (void)userContentController:(WKUserContentController *)userContentController
      didReceiveScriptMessage:(WKScriptMessage *)message {
	  //goLog("didReceiveScriptMessage: %s\n",[message.name UTF8String]);
    if ([message.name isEqualToString:@"wuapp"]) {
    	const char* str = [message.body UTF8String];
        WULog("Received event %s\n", str);
        receive(str);
        //(_GoString_){str, strlen(str)}
    }
}
@end


@interface WindowDelegate : NSObject <NSWindowDelegate> {
@private
    //NSView* _view;
}
@property (nonatomic, assign) NSView* view;
@end

@implementation WindowDelegate
//@synthesize view = _view;
- (void)windowDidResize:(NSNotification *)notification {
    //goUILog("windowDidResize\n");
}

- (void)windowDidMiniaturize:(NSNotification *)notification{
    //goUILog("windowDidMiniaturize\n");
}
- (void)windowDidEnterFullScreen:(NSNotification *)notification {
    //goUILog("windowDidEnterFullScreen\n");
}
- (void)windowDidExitFullScreen:(NSNotification *)notification {
    //goUILog("windowDidExitFullScreen\n");
}
- (void)windowDidBecomeKey:(NSNotification *)notification {
    //goUILog("Window: become key\n");
}

- (void)windowDidBecomeMain:(NSNotification *)notification {
    //goUILog("Window: become main\n");
}

- (void)windowDidResignKey:(NSNotification *)notification {
    //goUILog("Window: resign key\n");
}

- (void)windowDidResignMain:(NSNotification *)notification {
    //goUILog("Window: resign main\n");
}

- (void)windowWillClose:(NSNotification *)notification {
    [NSAutoreleasePool new];
    //goUILog("NSWindowDelegate::windowWillClose\n");
    [NSApp terminate:NSApp];
}
@end

@interface WUWindow : NSObject {
@private
    WKWebView* webView;
}
@end

@implementation WUWindow
//@synthesize webView = webView_;
//static WKWebView* webView;
- (void)create:(struct WindowSettings)settings {
    @autoreleasepool {
    	WindowDelegate* delegate = [[WindowDelegate alloc] init];
        NSRect rect = NSMakeRect(0, 0, settings.width, settings.height);
        id window = [[NSWindow alloc]
                     initWithContentRect:rect
                     styleMask:(NSWindowStyleMaskTitled |
                                NSWindowStyleMaskClosable |
                                NSWindowStyleMaskMiniaturizable |
                                NSWindowStyleMaskResizable |
                                NSWindowStyleMaskUnifiedTitleAndToolbar )
                     backing:NSBackingStoreBuffered
                     defer:NO];

        //delegate.view = [window contentView];
        [window setDelegate:delegate];
        [window cascadeTopLeftFromPoint:NSMakePoint(settings.left,settings.top)];
        [window setTitle:[NSString stringWithUTF8String:settings.title]];

        MessageHandler* handler = [[MessageHandler alloc] init];

        WKUserContentController *userContentController = [[WKUserContentController alloc] init];
        [userContentController addScriptMessageHandler:handler name:@"wuapp"];
        WKWebViewConfiguration *configuration = [[WKWebViewConfiguration alloc] init];
        configuration.userContentController = userContentController;

        webView = [[WKWebView alloc] initWithFrame:rect configuration:configuration];
        [webView.configuration.preferences setValue:@YES forKey:@"developerExtrasEnabled"];
		[webView.configuration.preferences setValue:@YES forKey:@"allowFileAccessFromFileURLs"];
		[webView.configuration setValue:@YES forKey:@"allowUniversalAccessFromFileURLs"];

        NSString *index = [NSString stringWithUTF8String:settings.index];
        if([index hasPrefix:@"http"]) {
        	NSURL *nsURL = [NSURL URLWithString:index];
        	NSURLRequest *requestObj = [NSURLRequest requestWithURL:nsURL];
			[webView loadRequest:requestObj];
        } else {
			NSString *bundlePath = [[NSBundle mainBundle] resourcePath];
			NSString *dir = [bundlePath stringByAppendingPathComponent:[NSString stringWithUTF8String:settings.webDir]];
			index = [dir stringByAppendingPathComponent:index];
			//goUILog("bundlePath:%s",[bundlePath UTF8String]);
			//goUILog("dir:%s",[dir UTF8String]);
			//goUILog("index:%s",[index UTF8String]);

			NSURL *nsURL = [NSURL fileURLWithPath:index isDirectory:false];
			NSURL *nsDir = [NSURL fileURLWithPath:dir isDirectory:true];
			[webView loadFileURL:nsURL allowingReadAccessToURL:nsDir];
		}

        [[window contentView] addSubview:webView];
        [window makeKeyAndOrderFront:nil];
    }
}

-(void) evaluateJS:(NSString*)script {
	printf("window evaluatejs:%p\n",script);
	const char* log = [script UTF8String];
printf("window evaluatejs log:%p\n",log);
    //goUILog("evalue:%s",log);
    [webView evaluateJavaScript:script completionHandler:^(id _Nullable response, NSError * _Nullable error) {
        //goUILog("response:%s,error:%s",[response UTF8String],[error UTF8String]);
    }];
}
@end

//app
@interface ApplicationDelegate : NSObject <NSApplicationDelegate> {
@private
    MenuDef* _menuDefs;
    int _menuCount;
}
@property (nonatomic, assign) struct MenuDef* menuDefs;
@property (assign) int menuCount;
@end

@implementation ApplicationDelegate
-(void)applicationWillFinishLaunching:(NSNotification *)aNotification
{
    //goUILog("applicationWillFinishLaunching\n");
    [NSApplication sharedApplication];
    WULog("_menuCount: %d\n",_menuCount);
    if(_menuCount!=0) {
    	[WUMenu buildMenu:_menuDefs count:_menuCount ];
    }
    [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
}

-(void)applicationDidFinishLaunching:(NSNotification *)notification
{
	//goUILog("applicationDidFinishLaunching\n");
    [NSApplication sharedApplication];

    [NSApp activateIgnoringOtherApps:YES];
}
@end

@interface WUApp : NSObject

@end

@implementation WUApp
static WUWindow* window;

+(void)initialize {}
+(void)start:(WindowSettings)settings menuDefs:(struct MenuDef[])menuDefs menuCount: (int)menuCount {
    @autoreleasepool {
        [NSApplication sharedApplication];
        [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];

        ApplicationDelegate* appDelegate = [[ApplicationDelegate alloc] init];
        appDelegate.menuDefs = menuDefs;
        appDelegate.menuCount = menuCount;
        WULog("menuCount: %d\n",menuCount);
        [NSApp setDelegate:appDelegate];

        window = [[WUWindow alloc] init];
        [window create:settings];
		//goUILog("run loop");
        [NSApp run];
        //run();
        }
}

+(void)evaluateJS:(NSString*)js {
WULog("app evaluatejs:%p\n",js);
	[window evaluateJS:js];
}

+(void)exit {
    [NSApp terminate:NSApp];
}
@end

void create(WindowSettings settings, MenuDef* menuDefs, int menuCount) {
	[WUApp start:settings menuDefs:menuDefs menuCount:menuCount];
}

void invokeJS(const char *js,int async) {
	NSString* script = [NSString stringWithUTF8String:js];
	if(async) {
		WULog("invokeJS async:%s",js);
		dispatch_async_and_wait(dispatch_get_main_queue(),^{
			[WUApp evaluateJS:script];
		});
	} else {
		WULog("script:%s\n",script);
		[WUApp evaluateJS:script];
	}
}

void invokeJSB(const char *js,int length, int async) {
	//goUILog("invokeJSB:%p",js);
	NSString* script = [NSString stringWithUTF8String:js];
	//NSString* script = [[NSString alloc]initWithBytes:js length:length encoding:NSUTF8StringEncoding];
	if(async) {
		//goUILog("invokeJS async:%s",js);
		dispatch_async_and_wait(dispatch_get_main_queue(),^{
			[WUApp evaluateJS:script];
		});
	} else {
		//goUILog("invokeJS script:%p",script);
		[WUApp evaluateJS:script];

	}
}

void exitApp() {
	[WUApp exit];
}

*/
import "C"

func cCreate(cs C.WindowSettings, cMenuDefs *C.MenuDef, count C.int) {
	C.create(cs, cMenuDefs, count)
}

func cActivate() {

}

func cInvokeJS(js *C.char, async C.int) {
	C.invokeJS(js, async)
}

func cExit() {
	C.exitApp()
}
