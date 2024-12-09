"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[1154],{6622:(e,n,s)=>{s.r(n),s.d(n,{assets:()=>o,contentTitle:()=>l,default:()=>d,frontMatter:()=>r,metadata:()=>a,toc:()=>c});var t=s(4848),i=s(8453);const r={sidebar_position:8},l="Packet captures",a={id:"examples/captures",title:"Packet captures",description:"Powered by PCAPng, WireShark, tshark",source:"@site/docs/examples/captures.md",sourceDirName:"examples",slug:"/examples/captures",permalink:"/examples/captures",draft:!1,unlisted:!1,editUrl:"https://github.com/cunicu/gont/edit/main/website/docs/examples/captures.md",tags:[],version:"current",sidebarPosition:8,frontMatter:{sidebar_position:8},sidebar:"docsSidebar",previous:{title:"Debugging",permalink:"/examples/debugging"},next:{title:"Network Emulation",permalink:"/examples/netem"}},o={},c=[{value:"Write to packets to",id:"write-to-packets-to",level:2},{value:"Filtering",id:"filtering",level:2},{value:"Session key logging",id:"session-key-logging",level:2},{value:"Example",id:"example",level:2}];function p(e){const n={a:"a",h1:"h1",h2:"h2",header:"header",img:"img",li:"li",p:"p",ul:"ul",...(0,i.R)(),...e.components};return(0,t.jsxs)(t.Fragment,{children:[(0,t.jsx)(n.header,{children:(0,t.jsx)(n.h1,{id:"packet-captures",children:"Packet captures"})}),"\n",(0,t.jsxs)(n.p,{children:["Powered by ",(0,t.jsx)(n.a,{href:"https://pcapng.com/",children:"PCAPng"}),", ",(0,t.jsx)(n.a,{href:"https://www.wireshark.org/",children:"WireShark"}),", ",(0,t.jsx)(n.a,{href:"https://www.wireshark.org/docs/man-pages/tshark.html",children:"tshark"})]}),"\n",(0,t.jsx)(n.h2,{id:"write-to-packets-to",children:"Write to packets to"}),"\n",(0,t.jsx)(n.p,{children:"Gont merges and sorts packet captures in real-time\nfrom multiple interfaces and records them to one of the following sinks:"}),"\n",(0,t.jsxs)(n.ul,{children:["\n",(0,t.jsxs)(n.li,{children:["Using ",(0,t.jsx)(n.a,{href:"https://github.com/pcapng/pcapng",children:"PCAPng"})," format","\n",(0,t.jsxs)(n.ul,{children:["\n",(0,t.jsx)(n.li,{children:"Regular files"}),"\n",(0,t.jsx)(n.li,{children:"Named pipes"}),"\n",(0,t.jsx)(n.li,{children:"TCP / UDP / Unix listeners"}),"\n",(0,t.jsx)(n.li,{children:"WireShark real-time stream"}),"\n"]}),"\n"]}),"\n",(0,t.jsx)(n.li,{children:"Go channels"}),"\n",(0,t.jsx)(n.li,{children:"Go callback functions."}),"\n"]}),"\n",(0,t.jsx)(n.h2,{id:"filtering",children:"Filtering"}),"\n",(0,t.jsx)(n.p,{children:"Captured network traffic can be filtered by"}),"\n",(0,t.jsxs)(n.ul,{children:["\n",(0,t.jsx)(n.li,{children:"Selected Gont nodes and interfaces"}),"\n",(0,t.jsx)(n.li,{children:"eBPF filter programs"}),"\n",(0,t.jsxs)(n.li,{children:[(0,t.jsx)(n.a,{href:"https://www.tcpdump.org/manpages/pcap-filter.7.html",children:"pcap-filter(7)"})," expressions"]}),"\n",(0,t.jsx)(n.li,{children:"Go callback functions (\u26a0 slow!)"}),"\n"]}),"\n",(0,t.jsx)(n.h2,{id:"session-key-logging",children:"Session key logging"}),"\n",(0,t.jsxs)(n.p,{children:["Most transport layer encryption protocols today provide ",(0,t.jsx)(n.a,{href:"https://en.wikipedia.org/wiki/Forward_secrecy",children:"perfect forward secrecy"})," by using short-lived ephemeral session keys."]}),"\n",(0,t.jsx)(n.p,{children:"Gont offers a feature to log these session keys into a PCAPng file to enable a decryption of upper layer protocols with a dissector tool like Wireshark."}),"\n",(0,t.jsx)(n.h2,{id:"example",children:"Example"}),"\n",(0,t.jsx)(n.p,{children:(0,t.jsx)(n.img,{src:s(4343).A+"",width:"2152",height:"1270"})})]})}function d(e={}){const{wrapper:n}={...(0,i.R)(),...e.components};return n?(0,t.jsx)(n,{...e,children:(0,t.jsx)(p,{...e})}):p(e)}},4343:(e,n,s)=>{s.d(n,{A:()=>t});const t=s.p+"assets/images/session-key-logging-16dac32bc665fa024f8f540adb9fb2b1.png"},8453:(e,n,s)=>{s.d(n,{R:()=>l,x:()=>a});var t=s(6540);const i={},r=t.createContext(i);function l(e){const n=t.useContext(r);return t.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function a(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(i):e.components||i:l(e.components),t.createElement(r.Provider,{value:n},e.children)}}}]);