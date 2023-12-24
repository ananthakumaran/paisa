// This file was generated by lezer-generator. You probably shouldn't edit it.
import {LRParser} from "@lezer/lr"
import {jsonHighlighting} from "./highlight"
export const parser = LRParser.deserialize({
  version: 14,
  states: "$bQVQPOOOOQO'#C`'#C`O!WQQO'#CdOOQO'#C_'#C_OOQO'#Cg'#CgO!]QPO'#CfOOQO'#C{'#C{O!qQPO'#DOOVQPO'#CwOVQPO'#C}OOQO'#C^'#C^QVQPOOOOQO'#DP'#DPO$OQQO,59OOOQO'#Cp'#CpO$WQPO,59QOOQO'#Cx'#CxOVQPO,59cOOQO,59c,59cO$iQPO,59iOOQO-E6|-E6|OOQO-E6}-E6}OOQO1G.j1G.jOOQO1G.l1G.lOOQO1G.}1G.}OOQO1G/T1G/T",
  stateData: "%P~OvOS~OTPOUPOVROXRO[SO]SO^SO_SO`SOaSObSOcSOpUOxQO{XO~Oy[O~Oe^Of^Og^Oh^Oi^Oj^O~Om`On`OTrXUrXVrXXrX[rX]rX^rX_rX`rXarXbrXcrXprXtrXxrX{rXzrX~OwfOy[O~OTPOUPOVROXROxQO~OziO~PVO[]^_`abcpmnTVb~",
  goto: "#gtPPu!R!^PPP!^P!g!oPPPPPPPP!wPPPPPP!g!zPP!}P!g#V#aWVOXZcQbWRha[YOWXZacRg__ROWXZ_ac]YOWXZac]TOWXZacR_TRaV]WOWXZacQZOQcXTdZcQ]QRe]",
  nodeNames: "⚠ Query Clause Value String UnQuoted Quoted Number DateValue RegExp Condition Property Account Commodity Amount Total Filename Note Payee Date Operator = =~ > >= < <= BooleanCondition BooleanBinaryOperator AND OR BooleanUnaryOperator NOT Expression",
  maxTerm: 43,
  propSources: [jsonHighlighting],
  skippedNodes: [0],
  repeatNodeCount: 2,
  tokenData: "!!_~RsXY#`YZ#`]^#`pq#`rs#exy%}yz&S}!O&X!O!P(T!P!Q(Y!Q!R+d!R![+t!^!_,X!_!`,f!`!a,s!c!d-Q!d!p/b!p!q/{!q!r1t!r!}/b!}#O3O#P#Q3T#T#U3Y#U#V/b#V#W:t#W#X@c#X#Y/b#Y#ZBv#Z#b/b#b#cGs#c#d/b#d#eJZ#e#h/b#h#iM]#i#o/b~#eOv~~#hWpq#eqr#ers$Qs#O#e#O#P$V#P;'S#e;'S;=`%w<%lO#e~$VOU~~$YXrs#e!P!Q#e#O#P#e#U#V#e#Y#Z#e#b#c#e#f#g#e#h#i#e#i#j$u~$xR!Q![%R!c!i%R#T#Z%R~%UR!Q![%_!c!i%_#T#Z%_~%bR!Q![%k!c!i%k#T#Z%k~%nR!Q![#e!c!i#e#T#Z#e~%zP;=`<%l#e~&SO{~~&XOz~R&^QyQ!Q!R&d!R!['rP&iRVP!O!P&r!g!h'W#X#Y'WP&uP!Q![&xP&}RVP!Q![&x!g!h'W#X#Y'WP'ZR{|'d}!O'd!Q!['jP'gP!Q!['jP'oPVP!Q!['jP'wSVP!O!P&r!Q!['r!g!h'W#X#Y'WQ(YOyQR(_WyQOY(wZ!P(w!Q!}(w!}#O*O#O#P*}#P;'S(w;'S;=`+^<%lO(wP(zXOY(wZ!P(w!P!Q)g!Q!}(w!}#O*O#O#P*}#P;'S(w;'S;=`+^<%lO(wP)lUXP#Z#[)g#]#^)g#a#b)g#g#h)g#i#j)g#m#n)gP*RVOY*OZ#O*O#O#P*h#P#Q(w#Q;'S*O;'S;=`*w<%lO*OP*kSOY*OZ;'S*O;'S;=`*w<%lO*OP*zP;=`<%l*OP+QSOY(wZ;'S(w;'S;=`+^<%lO(wP+aP;=`<%l(wR+kRyQVP!O!P&r!g!h'W#X#Y'WR+{SyQVP!O!P&r!Q!['r!g!h'W#X#Y'W~,^Pi~!_!`,a~,fOj~~,kPe~#r#s,n~,sOf~~,xPg~!_!`,{~-QOh~R-XWyQTP!O!P-q!P!Q-q!Q![-q![!]-q!c!p-q!p!q.Y!q!}-q#T#o-qP-vUTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#o-qP._WTP!O!P-q!P!Q-q!Q![-q![!]-q!c!f-q!f!g.w!g!}-q#T#o-qP/OUmPTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#o-qR/iUyQTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#o-qR0SWyQTP!O!P-q!P!Q-q!Q![-q![!]-q!c!q-q!q!r0l!r!}-q#T#o-qP0qWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!v-q!v!w1Z!w!}-q#T#o-qP1bUpPTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#o-qR1{WyQTP!O!P-q!P!Q-q!Q![-q![!]-q!c!t-q!t!u2e!u!}-q#T#o-qP2lUnPTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#o-q~3TOx~~3YOw~R3aYyQTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#V-q#V#W4P#W#a-q#a#b7q#b#o-qP4UWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#V-q#V#W4n#W#o-qP4sWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#c-q#c#d5]#d#o-qP5bWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#i-q#i#j5z#j#o-qP6PWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#b-q#b#c6i#c#o-qP6nWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#h-q#h#i7W#i#o-qP7_U[PTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#o-qP7vWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#c-q#c#d8`#d#o-qP8eWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#i-q#i#j8}#j#o-qP9SWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#b-q#b#c9l#c#o-qP9qWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#h-q#h#i:Z#i#o-qP:bU^PTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#o-qR:{WyQTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#c-q#c#d;e#d#o-qP;jWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#a-q#a#b<S#b#o-qP<XWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#a-q#a#b<q#b#o-qP<vWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#c-q#c#d=`#d#o-qP=eWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#W-q#W#X=}#X#o-qP>SWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#]-q#]#^>l#^#o-qP>qWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#h-q#h#i?Z#i#o-qP?`WTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#m-q#m#n?x#n#o-qP@PU]PTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#o-qR@jVyQTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#UAP#U#o-qPAUWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#h-q#h#iAn#i#o-qPAsWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#X-q#X#YB]#Y#o-qPBdUcPTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#o-qRB}WyQTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#]-q#]#^Cg#^#o-qPClWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#`-q#`#aDU#a#o-qPDZWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#X-q#X#YDs#Y#o-qPDxWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#b-q#b#cEb#c#o-qPEgVTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#UE|#U#o-qPFRWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#a-q#a#bFk#b#o-qPFpWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#X-q#X#YGY#Y#o-qPGaU`PTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#o-qRGzWyQTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#c-q#c#dHd#d#o-qPHiWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#h-q#h#iIR#i#o-qPIWWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#X-q#X#YIp#Y#o-qPIwUaPTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#o-qRJbVyQTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#UJw#U#o-qPJ|WTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#m-q#m#nKf#n#o-qPKkWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#X-q#X#YLT#Y#o-qPLYWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#X-q#X#YLr#Y#o-qPLyUbPTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#o-qRMdWyQTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#c-q#c#dM|#d#o-qPNRWTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#h-q#h#iNk#i#o-qPNpVTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#U! V#U#o-qP! [WTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#`-q#`#a! t#a#o-qP! {U_PTP!O!P-q!P!Q-q!Q![-q![!]-q!c!}-q#T#o-q",
  tokenizers: [0, 1],
  topRules: {"Query":[0,1]},
  tokenPrec: 170,
  termNames: {"0":"⚠","1":"@top","2":"Clause","3":"Value","4":"String","5":"UnQuoted","6":"Quoted","7":"Number","8":"DateValue","9":"RegExp","10":"Condition","11":"Property","12":"Account","13":"Commodity","14":"Amount","15":"Total","16":"Filename","17":"Note","18":"Payee","19":"Date","20":"Operator","21":"\"=\"","22":"\"=~\"","23":"\">\"","24":"\">=\"","25":"\"<\"","26":"\"<=\"","27":"BooleanCondition","28":"BooleanBinaryOperator","29":"AND","30":"OR","31":"BooleanUnaryOperator","32":"NOT","33":"Expression","34":"Clause+","35":"dateChar+","36":"␄","37":"%mainskip","38":"whitespace","39":"\"]\"","40":"\"[\"","41":"dateChar","42":"\")\"","43":"\"(\""}
})