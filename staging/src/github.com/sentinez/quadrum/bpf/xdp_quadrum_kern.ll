; ModuleID = 'xdp_quadrum_kern.c'
source_filename = "xdp_quadrum_kern.c"
target datalayout = "e-m:e-p:64:64-i64:64-i128:128-n32:64-S128"
target triple = "bpf"

%struct.xdp_md = type { i32, i32, i32, i32, i32, i32 }
%struct.ethhdr = type { [6 x i8], [6 x i8], i16 }

@LICENSE = dso_local global [4 x i8] c"GPL\00", section "license", align 1
@xdp_quadrum_prog.____fmt = internal constant [33 x i8] c"[quadrum] packet size: %d bytes\0A\00", align 1
@llvm.compiler.used = appending global [2 x ptr] [ptr @LICENSE, ptr @xdp_quadrum_prog], section "llvm.metadata"

; Function Attrs: nounwind
define dso_local noundef i32 @xdp_quadrum_prog(ptr nocapture noundef readonly %0) #0 section "xdp" {
  %2 = load i32, ptr %0, align 4, !tbaa !3
  %3 = zext i32 %2 to i64
  %4 = inttoptr i64 %3 to ptr
  %5 = getelementptr inbounds %struct.xdp_md, ptr %0, i64 0, i32 1
  %6 = load i32, ptr %5, align 4, !tbaa !8
  %7 = zext i32 %6 to i64
  %8 = inttoptr i64 %7 to ptr
  %9 = getelementptr inbounds %struct.ethhdr, ptr %4, i64 1
  %10 = icmp ugt ptr %9, %8
  %11 = getelementptr inbounds %struct.ethhdr, ptr %4, i64 2, i32 1
  %12 = icmp ugt ptr %11, %8
  %13 = select i1 %10, i1 true, i1 %12
  br i1 %13, label %17, label %14

14:                                               ; preds = %1
  %15 = sub i32 %6, %2
  %16 = tail call i64 (ptr, i32, ...) inttoptr (i64 6 to ptr)(ptr noundef nonnull @xdp_quadrum_prog.____fmt, i32 noundef 33, i32 noundef %15) #1
  br label %17

17:                                               ; preds = %14, %1
  %18 = phi i32 [ 0, %1 ], [ 2, %14 ]
  ret i32 %18
}

attributes #0 = { nounwind "frame-pointer"="all" "no-trapping-math"="true" "stack-protector-buffer-size"="8" }
attributes #1 = { nounwind }

!llvm.module.flags = !{!0, !1}
!llvm.ident = !{!2}

!0 = !{i32 1, !"wchar_size", i32 4}
!1 = !{i32 7, !"frame-pointer", i32 2}
!2 = !{!"Ubuntu clang version 18.1.3 (1ubuntu1)"}
!3 = !{!4, !5, i64 0}
!4 = !{!"xdp_md", !5, i64 0, !5, i64 4, !5, i64 8, !5, i64 12, !5, i64 16, !5, i64 20}
!5 = !{!"int", !6, i64 0}
!6 = !{!"omnipotent char", !7, i64 0}
!7 = !{!"Simple C/C++ TBAA"}
!8 = !{!4, !5, i64 4}
