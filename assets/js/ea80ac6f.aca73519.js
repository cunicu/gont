"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[8597],{7164:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>c,contentTitle:()=>r,default:()=>m,frontMatter:()=>i,metadata:()=>a,toc:()=>d});var o=n(4848),s=n(8453);const i={sidebar_position:10},r="Network Emulation",a={id:"examples/netem",title:"Network Emulation",description:"Powered by Linux's Traffic Control: Netem Qdisc.",source:"@site/docs/examples/netem.md",sourceDirName:"examples",slug:"/examples/netem",permalink:"/examples/netem",draft:!1,unlisted:!1,editUrl:"https://github.com/cunicu/gont/tree/main/docs/website/docs/examples/netem.md",tags:[],version:"current",sidebarPosition:10,frontMatter:{sidebar_position:10},sidebar:"docsSidebar",previous:{title:"Packet captures",permalink:"/examples/captures"},next:{title:"Firewall",permalink:"/examples/filtering"}},c={},d=[{value:"Attach a netem Qdisc to an interface",id:"attach-a-netem-qdisc-to-an-interface",level:2}];function l(e){const t={a:"a",code:"code",h1:"h1",h2:"h2",header:"header",p:"p",pre:"pre",...(0,s.R)(),...e.components};return(0,o.jsxs)(o.Fragment,{children:[(0,o.jsx)(t.header,{children:(0,o.jsx)(t.h1,{id:"network-emulation",children:"Network Emulation"})}),"\n",(0,o.jsxs)(t.p,{children:["Powered by Linux's Traffic Control: ",(0,o.jsx)(t.a,{href:"https://man7.org/linux/man-pages/man8/tc-netem.8.html",children:"Netem Qdisc"}),"."]}),"\n",(0,o.jsx)(t.h2,{id:"attach-a-netem-qdisc-to-an-interface",children:"Attach a netem Qdisc to an interface"}),"\n",(0,o.jsx)(t.pre,{children:(0,o.jsx)(t.code,{className:"language-go",children:'import tcopt "github.com/cunicu/gont/v2/options/tc"\n\nnetwork.AddLink(\n  gont.NewInterface("eth0", host1,\n    opt.WithNetem(\n      tcopt.Latency(50 * time.Millisecond),\n      tcopt.Jitter(5 * time.Millisecond),\n      tcopt.Loss(0.1),\n    ),\n    opt.AddressIP("10.0.0.1/24")),\n  gont.NewInterface("eth0", host2,\n    opt.AddressIP("10.0.0.2/24")),\n)\n\nhost1.Ping(host2)\n'})})]})}function m(e={}){const{wrapper:t}={...(0,s.R)(),...e.components};return t?(0,o.jsx)(t,{...e,children:(0,o.jsx)(l,{...e})}):l(e)}},8453:(e,t,n)=>{n.d(t,{R:()=>r,x:()=>a});var o=n(6540);const s={},i=o.createContext(s);function r(e){const t=o.useContext(i);return o.useMemo((function(){return"function"==typeof e?e(t):{...t,...e}}),[t,e])}function a(e){let t;return t=e.disableParentContext?"function"==typeof e.components?e.components(s):e.components||s:r(e.components),o.createElement(i.Provider,{value:t},e.children)}}}]);