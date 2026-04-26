/**
 * Image utility functions
 */

export async function imageSourceToFile(
  imageSource: string,
  filename: string
): Promise<File> {
  if (imageSource.startsWith("data:")) {
    const [header, base64 = ""] = imageSource.split(",");
    const mimeType = header.match(/data:(.*?);base64/)?.[1] || "image/png";
    const binary = window.atob(base64);
    const bytes = new Uint8Array(binary.length);

    for (let i = 0; i < binary.length; i += 1) {
      bytes[i] = binary.charCodeAt(i);
    }

    return new File([bytes], filename, { type: mimeType });
  }

  const response = await fetch(imageSource);
  if (!response.ok) {
    throw new Error(`Failed to load image: HTTP ${response.status}`);
  }

  const blob = await response.blob();
  return new File([blob], filename, { type: blob.type || "image/png" });
}

export function dataUrlToFile(dataUrl: string, filename: string): File {
  const [header, base64 = ""] = dataUrl.split(",");
  const mimeType = header.match(/data:(.*?);base64/)?.[1] || "image/png";
  const binary = window.atob(base64);
  const bytes = new Uint8Array(binary.length);

  for (let i = 0; i < binary.length; i += 1) {
    bytes[i] = binary.charCodeAt(i);
  }

  return new File([bytes], filename, { type: mimeType });
}
