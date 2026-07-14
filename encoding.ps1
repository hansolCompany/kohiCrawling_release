# Windows PowerShell 한글 출력/입력 UTF-8 설정
if ($PSVersionTable.PSVersion.Major -lt 6) {
    chcp 65001 | Out-Null
}

$utf8 = [System.Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding = $utf8
[Console]::InputEncoding = $utf8
$OutputEncoding = $utf8

if ($Host.Name -eq "ConsoleHost") {
    try {
        $Host.UI.RawUI.OutputEncoding = $utf8
    } catch {
        # 일부 호스트에서는 미지원
    }
}
