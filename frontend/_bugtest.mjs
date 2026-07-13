import { parseLrc, parseLrcA2 } from '@applemusic-like-lyrics/lyric'

function isBracketA2Format(lrc) {
  return /\[\d{2}:\d{2}(?:\.\d+)?\]\s*[^\s\[\]]+\s*\[\d{2}:\d{2}(?:\.\d+)?\]/m.test(lrc)
}
function convertBracketA2(lrc) {
  if (!lrc) return ''
  return lrc
    .split('\n')
    .map((line) => {
      const headMatch = line.match(/^\[\d{2}:\d{2}(?:\.\d+)?\]/)
      if (!headMatch) return line
      const headEnd = headMatch.index + headMatch[0].length
      const head = line.slice(0, headEnd)
      const tail = line.slice(headEnd)
      const convertedTail = tail.replace(
        /\[\d{2}:\d{2}(?:\.\d+)?\]/g,
        (t) => `<${t.slice(1, -1)}>`,
      )
      return head + convertedTail
    })
    .join('\n')
}

const bracketA2 = `[00:00.00] [00:00.04] When [00:00.16] the [00:00.82] truth
[00:04.50] [00:04.54] we [00:04.90] were [00:05.20] young`

console.log('=== isBracketA2Format(bracketA2) ===', isBracketA2Format(bracketA2))
const conv = convertBracketA2(bracketA2)
console.log('=== converted ===\n' + conv)
const parsedA2 = parseLrcA2(conv)
console.log('=== parseLrcA2 result ===')
console.log(JSON.stringify(parsedA2, null, 1))

// 普通 LRC 误判测试
const plain = `[00:12.34] 普通歌词第一行
[00:18.00] 第二行有数字 2024 年 [02:00] 这样
[00:25.00] 第三行`
console.log('\n=== isBracketA2Format(plain) ===', isBracketA2Format(plain))
console.log('=== converted plain ===\n' + convertBracketA2(plain))
console.log('=== parseLrcA2(plain converted) ===')
console.log(JSON.stringify(parseLrcA2(convertBracketA2(plain)), null, 1))
console.log('=== parseLrc(plain) ===')
console.log(JSON.stringify(parseLrc(plain), null, 1))
