"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[4101],{45:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>d,contentTitle:()=>r,default:()=>g,frontMatter:()=>a,metadata:()=>s,toc:()=>c});var o=t(4848),i=t(8453);const a={sidebar_position:7},r="Debugging",s={id:"examples/debugging",title:"Debugging",description:"Powered by Delve Debugger.",source:"@site/docs/examples/debugging.md",sourceDirName:"examples",slug:"/examples/debugging",permalink:"/examples/debugging",draft:!1,unlisted:!1,editUrl:"https://github.com/cunicu/gont/edit/main/website/docs/examples/debugging.md",tags:[],version:"current",sidebarPosition:7,frontMatter:{sidebar_position:7},sidebar:"docsSidebar",previous:{title:"Tracing",permalink:"/examples/tracing"},next:{title:"Packet captures",permalink:"/examples/captures"}},d={},c=[{value:"Introduction",id:"introduction",level:2},{value:"Create a debugger",id:"create-a-debugger",level:2},{value:"Attach a debugger",id:"attach-a-debugger",level:2},{value:"Define tracepoints",id:"define-tracepoints",level:2},{value:"Break-, watchpoint location",id:"break--watchpoint-location",level:3},{value:"Define tracepoints",id:"define-tracepoints-1",level:2},{value:"Gathering of breakpoint information",id:"gathering-of-breakpoint-information",level:3},{value:"VSCode Integration",id:"vscode-integration",level:2}];function l(e){const n={a:"a",code:"code",h1:"h1",h2:"h2",h3:"h3",header:"header",p:"p",pre:"pre",...(0,i.R)(),...e.components};return(0,o.jsxs)(o.Fragment,{children:[(0,o.jsx)(n.header,{children:(0,o.jsx)(n.h1,{id:"debugging",children:"Debugging"})}),"\n",(0,o.jsxs)(n.p,{children:["Powered by ",(0,o.jsx)(n.a,{href:"https://github.com/go-delve/delve",children:"Delve Debugger"}),"."]}),"\n",(0,o.jsx)(n.h2,{id:"introduction",children:"Introduction"}),"\n",(0,o.jsx)(n.p,{children:"Gont can manage Delve debugger instances attached to spawned\nsub-processes. These Delve instances can be used to debug or trace\nmultiple applications in a distributed system simultaneously. Users can\neither use DAP-compatible IDEs like VScode to attach to those processes,\nor record tracepoint data generated by Delve break- & watchpoints to a\nPCAPng file."}),"\n",(0,o.jsx)(n.p,{children:"Its integration with the packet capture and event tracing feature of\nGont allows the user to even streaming tracepoint data interleaved with\npacket and other tracing data in real-time to Wireshark."}),"\n",(0,o.jsx)(n.h2,{id:"create-a-debugger",children:"Create a debugger"}),"\n",(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-go",children:'import dopt "cunicu.li/gont/v2/pkt/options/debug"\n\nt := gont.NewTracer(...)\n\nd := gont.NewDebugger(\n  // ... Tracepoints are defined here\n  dopt.BreakOnEntry(true),\n  dopt.ToTracer(t),\n\n  dopt.ListenAddr("tcp:[::]:1234")) // Listening socket for connection of external DAP client\n'})}),"\n",(0,o.jsx)(n.h2,{id:"attach-a-debugger",children:"Attach a debugger"}),"\n",(0,o.jsx)(n.p,{children:"Debug all processes started by nodes of this network:"}),"\n",(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-go",children:'network, _ := gont.NewNetwork("", d)\n'})}),"\n",(0,o.jsx)(n.p,{children:"Debug all processes started by a node:"}),"\n",(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-go",children:'host1 := network.NewHost("host1", d)\n'})}),"\n",(0,o.jsx)(n.p,{children:"Debug a single process:"}),"\n",(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-go",children:'host1.RunGo("test/main.go", d)\n'})}),"\n",(0,o.jsx)(n.p,{children:"(Like for the event tracing)"}),"\n",(0,o.jsx)(n.h2,{id:"define-tracepoints",children:"Define tracepoints"}),"\n",(0,o.jsx)(n.h3,{id:"break--watchpoint-location",children:"Break-, watchpoint location"}),"\n",(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-go",children:'import "github.com/go-delve/delve/service/api"\nimport dopt "cunicu.li/gont/v2/pkg/options/debug"\n\nd := gont.NewDebugger(\n  gont.NewTracepoint(\n    dopt.Disabled(false),\n    dopt.Name("tp1"),\n    dopt.Message("A trace message with evaluated {placeholders}"),\n    dopt.Location(...), // A Delve locspec\n    dopt.Address(0x12312321),\n    dopt.File("main.go"),\n    dopt.Line(12),\n    dopt.FunctionName("main.main"),\n    dopt.FunctionNameRegex("main\\.(main|setupLogger)"),\n    dopt.Condition("i % 100 == 0"),\n    dopt.HitCondition("> 100"),\n    dopt.HitConditionPerGoroutine(true),\n    dopt.Watch("p", api.WatchRead|api.WatchWrite)),\n  ...\n)\n'})}),"\n",(0,o.jsx)(n.h2,{id:"define-tracepoints-1",children:"Define tracepoints"}),"\n",(0,o.jsx)(n.h3,{id:"gathering-of-breakpoint-information",children:"Gathering of breakpoint information"}),"\n",(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-go",children:'import dopt "cunicu.li/gont/v2/pkt/options/debug"\n\nd := gont.NewDebugger(\n  gont.NewTracepoint(\n    ...\n    dopt.Variable("j"),\n    dopt.Goroutine(true),\n    dopt.Stacktrace(10),\n    dopt.LoadLocals(...),\n    dopt.LoadArguments(\n      dopt.FollowPointers(true),\n      dopt.MaxVariableRecurse(3),\n      dopt.MaxStringLen(128),\n      dopt.MaxArrayValues(128),\n      dopt.MaxStructFields(32))))\n'})}),"\n",(0,o.jsx)(n.h2,{id:"vscode-integration",children:"VSCode Integration"}),"\n",(0,o.jsx)(n.p,{children:"Gont generates VS Code launch compound configurations based on the\nactive debugging sessions."}),"\n",(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-go",children:'d.WriteVSCodeConfigs("", false)\n'})})]})}function g(e={}){const{wrapper:n}={...(0,i.R)(),...e.components};return n?(0,o.jsx)(n,{...e,children:(0,o.jsx)(l,{...e})}):l(e)}},8453:(e,n,t)=>{t.d(n,{R:()=>r,x:()=>s});var o=t(6540);const i={},a=o.createContext(i);function r(e){const n=o.useContext(a);return o.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function s(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(i):e.components||i:r(e.components),o.createElement(a.Provider,{value:n},e.children)}}}]);