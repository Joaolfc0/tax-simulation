import type { Config } from "tailwindcss";

export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: "rgb(68, 68, 68)",
        "ink-soft": "rgb(102, 102, 102)",
        "ink-faint": "rgb(150, 150, 150)",
        line: "#e5e5e5",
        muted: "#f5f5f5",
      },
      fontFamily: {
        sans: [
          "Inter",
          "ui-sans-serif",
          "system-ui",
          "-apple-system",
          "Segoe UI",
          "Roboto",
          "sans-serif",
        ],
      },
      boxShadow: {
        soft: "0 1px 2px rgba(0,0,0,0.04)",
      },
    },
  },
  plugins: [],
} satisfies Config;
