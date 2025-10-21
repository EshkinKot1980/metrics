## Результат оптимизации для Агента
```bash
go tool pprof -top -sample_index=alloc_space -diff_base=agent-base.pprof agent-result.pprof
##### output
#File: main
#Type: alloc_space
#Time: Oct 21, 2025 at 7:03am (MSK)
#Duration: 240.01s, Total samples = 57155.94kB 
#Showing nodes accounting for -4190.48kB, 7.33% of 57155.94kB total
#Dropped 9 nodes (cum <= 285.78kB)
#      flat  flat%   sum%        cum   cum%
#-3837.73kB  6.71%  6.71% -3837.73kB  6.71%  compress/flate.(*compressor).initDeflate (inline)
#-2060.02kB  3.60% 10.32% -2060.02kB  3.60%  github.com/EshkinKot1980/metrics/internal/agent/#monitor.(*Monitor).collectMemStats
# 2050.67kB  3.59%  6.73%  2050.67kB  3.59%  strings.(*Builder).grow
# 1195.29kB  2.09%  4.64% -3667.09kB  6.42%  compress/flate.(*compressor).init
#-1027.13kB  1.80%  6.44%  1536.73kB  2.69%  github.com/EshkinKot1980/metrics/internal/agent/#monitor.(*AdditionalMonitor).Poll
# -513.31kB   0.9%  7.33%  -513.31kB   0.9%  encoding/json.Marshal
#  513.12kB   0.9%  6.44%   513.12kB   0.9%  compress/flate.(*huffmanEncoder).generate
#  512.88kB   0.9%  5.54%  1539.28kB  2.69%  github.com/shirou/gopsutil/v4/cpu.TimesWithContext
# -512.56kB   0.9%  6.44% -1024.65kB  1.79%  compress/flate.newHuffmanBitWriter (inline)
#  512.10kB   0.9%  5.54%   512.10kB   0.9%  github.com/EshkinKot1980/metrics/internal/agent/storage.(*MemoryStorage).Put
# -512.09kB   0.9%  6.44%  -512.09kB   0.9%  compress/flate.newHuffmanEncoder (inline)
# -512.02kB   0.9%  7.33%  -512.02kB   0.9%  internal/profile.(*Profile).postDecode
#  512.01kB   0.9%  6.44% -4350.56kB  7.61%  github.com/EshkinKot1980/metrics/internal/agent/client.(*HTTPClient).Report
# -512.01kB   0.9%  7.33%  -512.01kB   0.9%  time.NewTimer
#    0.31kB 0.00055%  7.33%  2050.99kB  3.59%  github.com/shirou/gopsutil/v4/internal/common.ReadLinesOffsetN
#...
```
## Для сервера результат оптимизации можно увидеть только на графе(через браузер)

Для сервера большую часть ресурсов отжирает работа с БД, следом за ней идет работа с ФС(если она есть),
дальше удут gzip, encodin/json и сам профилирвкщик.
Это, конечно, можно оптимизировать, но четверть спринта для этого мало.