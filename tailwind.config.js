/** @type {import('tailwindcss').Config} */
export default {
  content: ["./internal/templates/**/*.templ"],
  plugins: [require("@tailwindcss/typography")],
  theme: {
    fontFamily: {
      excali: ["Excalifont", "sans-serif"],
      mono: ["SFMono-Regular", "jetbrains mono", "monospace"],
    },
  },
};
