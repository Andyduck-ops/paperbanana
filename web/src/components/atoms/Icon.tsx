import { memo, type SVGAttributes } from "react";

export type IconSize = "xs" | "sm" | "md" | "lg" | "xl";
export type IconColor = "inherit" | "current" | "primary" | "success" | "warning" | "danger" | "muted";

export interface IconProps extends SVGAttributes<SVGSVGElement> {
  name: string;
  size?: IconSize;
  color?: IconColor;
  className?: string;
}

const sizeMap: Record<IconSize, number> = {
  xs: 12,
  sm: 16,
  md: 20,
  lg: 24,
  xl: 32,
};

const colorMap: Record<IconColor, string> = {
  inherit: "inherit",
  current: "currentColor",
  primary: "var(--color-primary)",
  success: "var(--color-success)",
  warning: "var(--color-warning)",
  danger: "var(--color-danger)",
  muted: "var(--color-muted-foreground)",
};

// Icon registry - common icons
const icons: Record<string, string> = {
  // Navigation
  home: "M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2V9z M9 22V12h6v10",
  settings: "M12 15a3 3 0 100-6 3 3 0 000 6z M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-2 2 2 2 0 01-2-2v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83 0 2 2 0 010-2.83l.06-.06a1.65 1.65 0 00.33-1.82 1.65 1.65 0 00-1.51-1H3a2 2 0 01-2-2 2 2 0 012-2h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 010-2.83 2 2 0 012.83 0l.06.06a1.65 1.65 0 001.82.33H9a1.65 1.65 0 001-1.51V3a2 2 0 012-2 2 2 0 012 2v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 0 2 2 0 010 2.83l-.06.06a1.65 1.65 0 00-.33 1.82V9a1.65 1.65 0 001.51 1H21a2 2 0 012 2 2 2 0 01-2 2h-.09a1.65 1.65 0 00-1.51 1z",
  user: "M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2 M12 11a4 4 0 100-8 4 4 0 000 8z",
  menu: "M3 12h18 M3 6h18 M3 18h18",
  close: "M18 6L6 18 M6 6l12 12",
  
  // Actions
  plus: "M12 5v14 M5 12h14",
  minus: "M5 12h14",
  edit: "M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7 M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z",
  trash: "M3 6h18 M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2 M10 11v6 M14 11v6",
  check: "M20 6L9 17l-5-5",
  
  // Content
  search: "M11 19a8 8 0 100-16 8 8 0 000 16z M21 21l-4.35-4.35",
  image: "M21 19V5a2 2 0 00-2-2H5a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2z M8.5 10a1.5 1.5 0 100-3 1.5 1.5 0 000 3z M21 15l-5-5L5 21",
  download: "M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4 M7 10l5 5 5-5 M12 15V3",
  upload: "M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4 M17 8l-5-5-5 5 M12 3v12",
  copy: "M16 4h2a2 2 0 012 2v14a2 2 0 01-2 2H6a2 2 0 01-2-2V6a2 2 0 012-2h2 M8 2h8a2 2 0 012 2v2a2 2 0 01-2 2H8a2 2 0 01-2-2V4a2 2 0 012-2z",
  
  // Status
  info: "M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z M12 16v-4 M12 8h.01",
  warning: "M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z M12 9v4 M12 17h.01",
  error: "M12 8v4 M12 16h.01 M21 12a9 9 0 11-18 0 9 9 0 0118 0z",
  loading: "M12 2v4m0 12v4M4.93 4.93l2.83 2.83m8.48 8.48l2.83 2.83M2 12h4m12 0h4M4.93 19.07l2.83-2.83m8.48-8.48l2.83-2.83",
  
  // Arrows
  chevronLeft: "M15 18l-6-6 6-6",
  chevronRight: "M9 18l6-6-6-6",
  chevronUp: "M18 15l-6-6-6 6",
  chevronDown: "M6 9l6 6 6-6",
  arrowLeft: "M19 12H5 M12 19l-7-7 7-7",
  arrowRight: "M5 12h14 M12 5l7 7-7 7",
  
  // Misc
  sparkle: "M12 3L14.5 8.5L20 11L14.5 13.5L12 19L9.5 13.5L4 11L9.5 8.5L12 3Z",
  magic: "M21 16V8a2 2 0 00-1-1.73l-7-4a2 2 0 00-2 0l-7 4A2 2 0 003 8v8a2 2 0 001 1.73l7 4a2 2 0 002 0l7-4A2 2 0 0021 16z M12 22V12 M12 12L3 7 M12 12l9-5",
  history: "M3 3v5h5 M3.05 13A9 9 0 106 5.3L3 8 M12 7v5l4 2",
  folder: "M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z",
  file: "M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z M14 2v6h6 M16 13H8 M16 17H8 M10 9H8",
};

export const Icon = memo(function Icon({
  name,
  size = "md",
  color = "current",
  className = "",
  ...props
}: IconProps) {
  const path = icons[name];
  if (!path) {
    console.warn(`Icon "${name}" not found in registry`);
    return null;
  }

  return (
    <svg
      width={sizeMap[size]}
      height={sizeMap[size]}
      viewBox="0 0 24 24"
      fill="none"
      stroke={colorMap[color]}
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      {...props}
    >
      <path d={path} />
    </svg>
  );
});

Icon.displayName = "Icon";

// Convenience exports for common icons
export const IconHome = (props: Omit<IconProps, "name">) => <Icon name="home" {...props} />;
export const IconSettings = (props: Omit<IconProps, "name">) => <Icon name="settings" {...props} />;
export const IconPlus = (props: Omit<IconProps, "name">) => <Icon name="plus" {...props} />;
export const IconClose = (props: Omit<IconProps, "name">) => <Icon name="close" {...props} />;
export const IconCheck = (props: Omit<IconProps, "name">) => <Icon name="check" {...props} />;
export const IconSearch = (props: Omit<IconProps, "name">) => <Icon name="search" {...props} />;
export const IconImage = (props: Omit<IconProps, "name">) => <Icon name="image" {...props} />;
export const IconDownload = (props: Omit<IconProps, "name">) => <Icon name="download" {...props} />;
export const IconUpload = (props: Omit<IconProps, "name">) => <Icon name="upload" {...props} />;
export const IconCopy = (props: Omit<IconProps, "name">) => <Icon name="copy" {...props} />;
export const IconTrash = (props: Omit<IconProps, "name">) => <Icon name="trash" {...props} />;
export const IconEdit = (props: Omit<IconProps, "name">) => <Icon name="edit" {...props} />;
export const IconChevronLeft = (props: Omit<IconProps, "name">) => <Icon name="chevronLeft" {...props} />;
export const IconChevronRight = (props: Omit<IconProps, "name">) => <Icon name="chevronRight" {...props} />;
export const IconInfo = (props: Omit<IconProps, "name">) => <Icon name="info" {...props} />;
export const IconWarning = (props: Omit<IconProps, "name">) => <Icon name="warning" {...props} />;
export const IconError = (props: Omit<IconProps, "name">) => <Icon name="error" {...props} />;
export const IconLoading = (props: Omit<IconProps, "name">) => <Icon name="loading" {...props} />;
export const IconSparkle = (props: Omit<IconProps, "name">) => <Icon name="sparkle" {...props} />;
export const IconMagic = (props: Omit<IconProps, "name">) => <Icon name="magic" {...props} />;
export const IconHistory = (props: Omit<IconProps, "name">) => <Icon name="history" {...props} />;
export const IconFolder = (props: Omit<IconProps, "name">) => <Icon name="folder" {...props} />;
export const IconFile = (props: Omit<IconProps, "name">) => <Icon name="file" {...props} />;
