package processor

import (
	"math/rand"
	"time"
)

var matchMessages = []string{
	"✅ Everything matches. For once.",
	"😌 All receipts balanced. Miracles happen.",
	"👍 Beer and bottles are in sync. Cool.",
	"🎯 Perfect match. Even the bottles agree.",
	"😅 No mismatches. I won't get used to this.",
	"🧾 All receipts match. Boring, but correct.",
	"🍺 Counts align. The bottles behaved.",
	"✨ Clean run. Nothing to complain about.",
	"🙃 All matched. I'm almost disappointed.",
	"✅ Balanced. The math did its job.",
	"😎 Match confirmed. You may proceed.",
	"🟢 No issues. The universe is aligned.",
	"👌 All good. No drama today.",
	"🥱 Everything matches. Wake me when it doesn't.",
	"✅ Checks out. Go brag to someone.",
	"🍺🧴 Beer equals bottles. Who would’ve thought.",
	"🏁 Done. All receipts are clean.",
	"✅ Zero mismatches. Don’t change anything.",
	"😇 Perfect. I'll pretend this is normal.",
	"🧠 Math works today. Shocking.",
	"🟢 All green. The bottles are honest.",
	"✅ Matching totals across the board.",
	"😌 Smooth sailing. No mismatches found.",
	"✅ Everything lines up. Nice.",
	"🎉 All receipts match. Party postponed.",
	"✅ Looks good. Move along.",
	"🧾 Clean receipts. No issues.",
	"✅ No mismatches. You got lucky.",
	"😄 All matched. Nothing to fix.",
	"✅ Perfect balance. Like a zen garden.",
	"🟢 All good. Even the decimals behaved.",
	"✅ Beer vs bottles: tie game.",
	"😏 All matched. I wanted to complain, but I can’t.",
	"✅ Everything matches. I checked. Twice.",
	"🍻 Totals match. Raise a glass.",
	"✅ Green lights only.",
	"😌 Nothing to see here. All good.",
	"✅ Verified. No mismatches.",
	"🎯 Nailed it. Every receipt matches.",
	"✅ Nice. Clean and tidy.",
	"🧾 All receipts are balanced. Yawn.",
	"✅ No surprises. That’s a win.",
	"😎 Everything matches. I'll allow it.",
	"✅ Spotless. Zero mismatches.",
	"🟢 All checks passed.",
	"✅ The bottles and beer finally agree.",
	"😌 Balanced receipts. A rare sight.",
	"✅ Nothing broken. Move on.",
	"🧾 All good. Keep it up.",
	"✅ All matched. I guess today is fine.",
}

var mismatchMessages = []string{
	"⚠️ Mismatch detected. Obviously.",
	"😑 The bottles and beer disagree. Again.",
	"🙄 Totals don’t match. Shocking.",
	"⚠️ Something’s off. The math is unimpressed.",
	"😬 Mismatches found. Try not to cry.",
	"🤦 Bottles and beer can’t get along.",
	"⚠️ Discrepancy alert. I did the math.",
	"😏 Mismatch season is here.",
	"🧾 Not all receipts match. Surprise.",
	"⚠️ Found mismatches. I hope you like puzzles.",
	"😒 The numbers are arguing.",
	"⚠️ Beer and bottles aren’t friends today.",
	"🤷 Mismatches found. What did you expect?",
	"⚠️ Totals diverged. Reality hurts.",
	"😬 Some receipts are off. Obviously.",
	"⚠️ Mismatch detected. Please clap.",
	"😑 The bottles lied.",
	"⚠️ Beer math failed. Again.",
	"🙃 Not matching. I’ll wait.",
	"⚠️ Inconsistent totals. Math is mad.",
	"😏 Found mismatches. Happy now?",
	"⚠️ The balance is broken.",
	"😬 Receipts don’t line up. Fun.",
	"⚠️ Mismatch list incoming.",
	"😒 Bottles vs beer: not a love story.",
	"⚠️ Discrepancies detected. I’m not surprised.",
	"🙄 Totals are off. Classic.",
	"⚠️ Found issues. I did my part.",
	"😑 The numbers refuse to cooperate.",
	"⚠️ Mismatch detected. Again. Yes, again.",
	"😬 Beer and bottles are out of sync.",
	"⚠️ Errors found. Please pretend to care.",
	"🤦 Receipts failed the vibe check.",
	"⚠️ Totals don’t match. Big surprise.",
	"😒 Discrepancy report: incoming.",
	"⚠️ Mismatch alert. I can’t unsee it.",
	"🙃 The math is wrong. Not my fault.",
	"⚠️ Mismatches found. Details below.",
	"😑 Receipts are messy. Shocker.",
	"⚠️ Totals disagree. Again.",
	"😬 Beer vs bottles: mismatch edition.",
	"⚠️ Found some chaos.",
	"🙄 Something doesn’t add up. Literally.",
	"⚠️ The bottles are freelancing.",
	"😒 Not all receipts match. Cute.",
	"⚠️ Mismatch count > 0. Good luck.",
	"🤷 Numbers are off. Fix it maybe.",
	"⚠️ The math is not mathing.",
	"😑 Mismatches found. I’ll wait.",
	"⚠️ Balance is broken. Details below.",
}

var snarkMatchMessages = []string{
	"🎉 Everything matched. I’m almost proud. Almost.",
	"😌 All good. I tried to find a problem. There wasn’t one.",
	"✅ Clean receipts. I guess you’re doing your job today.",
	"🧠 Totals matched. The math gods accepted your offering.",
	"👌 Everything lines up. Enjoy this moment before reality returns.",
	"😏 Perfect match. Try not to ruin it in the next file.",
	"🍺🧴 Bottles and beer agree. I’ll pretend this is normal.",
	"✅ No mismatches. I checked twice just to be annoyed.",
	"🟢 All green. I’m bored now.",
	"😌 It matches. You can stop sweating for five minutes.",
}

var snarkMismatchMessages = []string{
	"🙃 Here we go again. The numbers are doing their own thing.",
	"😑 Surprise, another mismatch. It's like a hobby.",
	"⚠️ You had one job: make totals match. And yet.",
	"🤦 The bottles and beer are in a toxic relationship again.",
	"😏 I found mistakes. You’re welcome.",
	"⚠️ Mismatches detected. Please act shocked.",
	"😬 The math is screaming quietly in the corner.",
	"🙄 Totals are off. Classic.",
	"⚠️ This report comes with free disappointment.",
	"😒 The receipts failed the vibe check. Again.",
	"🤷 I checked. The totals did not.",
	"😬 Your numbers are freelancing.",
	"⚠️ Another mismatch. At this point, it’s tradition.",
	"🙃 You were close. Not close enough.",
	"😑 The math is fine. The data isn't.",
	"⚠️ Totals diverged. Reality continues to disappoint.",
	"😏 Beer and bottles have trust issues.",
	"🤦 You lost the plot somewhere between liters and bottles.",
	"⚠️ The bottles and beer are not on speaking terms.",
	"😒 This is why we can’t have nice things.",
	"🙄 If mismatches were a sport, you’d medal.",
	"⚠️ The counts are arguing again. Loudly.",
	"😬 Receipts are messy. Clean it up?",
	"🤷 I did the math. You did… something else.",
	"⚠️ Not matching. Not surprising.",
	"😑 Another mismatch. I’m tired.",
	"🙃 The totals took a scenic detour.",
	"⚠️ The numbers are off. They knew what they were doing.",
	"😏 Even the bottles rolled their eyes.",
	"🤦 Mismatch confirmed. Pretend to be surprised.",
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomMessage(messages []string) string {
	if len(messages) == 0 {
		return ""
	}
	return messages[rand.Intn(len(messages))]
}

func randomMatchMessage() string {
	return randomMessage(matchMessages)
}

func randomMismatchMessage() string {
	return randomMessage(mismatchMessages)
}

func randomSnark(match bool) string {
	if match {
		return randomMessage(snarkMatchMessages)
	}
	return randomMessage(snarkMismatchMessages)
}
