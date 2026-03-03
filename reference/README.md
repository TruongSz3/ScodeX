# Reference Workspace Guide

## Mục tiêu
- Cung cấp điểm vào nhanh để tra cứu và phân phối task cho subagents.
- Giữ quy trình evidence-first, tránh suy đoán.

## Quy tắc cốt lõi (bắt buộc)
1. Luôn load trước `reference/repomix-codex.xml` và `reference/repomix-opencode.xml` để map context.
2. Luôn spawn subagents để tra cứu/tìm hiểu, không kết luận từ 1 agent đơn lẻ.
3. Luôn dùng `reference/codex/` và `reference/opencode/` làm source tài liệu chính.

## Luồng nhanh
1. Nạp 2 repomix để dựng bản đồ cấu trúc tổng thể.
2. Chia task theo folder ownership.
3. Giao subagents tra cứu trong 2 source chính.
4. Tổng hợp, cross-check với repomix, rồi mới chốt hướng triển khai.
5. Verify theo package scope bị ảnh hưởng.

## Chi tiết vận hành
- Xem `reference/AGENTS.md` để dùng policy đầy đủ (read-only repomix, ownership, verify, reporting, branch).
