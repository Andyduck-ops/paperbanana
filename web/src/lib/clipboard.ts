export async function copyImageToClipboard(imageData: string): Promise<boolean> {
  try {
    const response = await fetch(`data:image/png;base64,${imageData}`);
    const blob = await response.blob();
    await navigator.clipboard.write([
      new ClipboardItem({ 'image/png': blob })
    ]);
    return true;
  } catch {
    return false;
  }
}
