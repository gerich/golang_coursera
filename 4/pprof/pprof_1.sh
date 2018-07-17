curl http://127.0.0.1:8080/debug/pprof/heap -o mem_out.txt
curl http://127.0.0.1:8080/debug/pprof/profile?seconds=5 -o cpu_out.txt

go tool pprof -svg -inuse_space pprof1 mem_out.txt > mem_is.svg
go tool pprof -svg -inuse_objects pprof1 mem_out.txt > mem_oo.svg
go tool pprof -svg -alloc_space pprof1 mem_out.txt > mem_as.svg
go tool pprof -svg -alloc_objects pprof1 mem_out.txt > mem_ao.svg
go tool pprof -svg pprof1 cpu_out.txt > cpu.svg