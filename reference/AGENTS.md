# AGENTS Playbook (Reference)

## Mục tiêu
- Đồng bộ cách làm việc của agents/subagents khi tra cứu và phân phối công việc.
- Tăng độ chính xác bằng quy trình evidence-first, giảm rủi ro suy diễn sai.

## Quy tắc bắt buộc
1. Bước đầu tiên mỗi phiên phải load:
   - `reference/repomix-codex.xml`
   - `reference/repomix-opencode.xml`
2. Repomix files là read-only; chỉ dùng để map/cross-reference, không chỉnh sửa trực tiếp.
3. Luôn spawn subagents để tra cứu/tìm hiểu; không dựa vào kết luận của 1 agent duy nhất.
4. Nguồn tài liệu chính bắt buộc:
   - `reference/codex/`
   - `reference/opencode/`
   - `reference/docs/` (nguồn tổng hợp/audit để tra cứu nhanh; không thay thế source gốc khi cần verify sâu)
5. Phân công ownership theo folder khi chia subagents; tuyệt đối tránh overlap chỉnh sửa file.
6. Chạy lệnh/test theo package scope phù hợp với vùng ảnh hưởng, tránh chạy tràn toàn workspace.
7. Nhánh mặc định của repo hiện tại: `main`.
8. Mỗi subagent phải trả report ngắn: scope đã đọc, kết luận chính, paths liên quan, unresolved questions.

## Quy trình đề xuất
1. Nạp repomix để dựng bản đồ context tổng thể.
2. Chia domain theo folder/package, gán owner rõ cho từng subagent.
3. Subagents tra cứu chi tiết trong `reference/codex/` và `reference/opencode/`; dùng `reference/docs/` làm baseline tổng hợp khi có.
4. Tổng hợp kết quả và cross-check lại với repomix + `reference/docs/` trước khi chốt quyết định.
5. Thực thi/chỉnh sửa trong scope ownership đã phân công.
6. Verify bằng test/lệnh đúng package scope; báo cáo kết quả theo format ngắn.

## Nguồn tham chiếu
- `reference/repomix-codex.xml`
- `reference/repomix-opencode.xml`
- `reference/codex/`
- `reference/opencode/`
- `reference/docs/`
