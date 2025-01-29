import EditorJS from "@editorjs/editorjs";
import Header from "@editorjs/header";
import Quote from "@editorjs/quote";
import EditorjsList from "@editorjs/list";
import DragDrop from "editorjs-drag-drop";
import Undo from "editorjs-undo";
import Table from "@editorjs/table";
import Paragraph from "editorjs-paragraph-with-alignment";
// import LinkTool from "@editorjs/link";
const EJLaTeX = require("editorjs-latex");
const editor = new EditorJS({
  holder: "editor",
  onReady: () => {
    new Undo({ editor });
    new DragDrop(editor);
  },
  tools: {
    Math: {
      class: EJLaTeX,
      shortcut: "CMD+SHIFT+L",
      inlineToolbar: true,
      config: {
        css: ".math-input-wrapper { padding: 5px; }",
      },
    },
    //TODO: setup backend for fetch point
    // linkTool: {
    //   class: LinkTool,
    //   config: {
    //     endpoint: "http://localhost:8008/fetchUrl", // Your backend endpoint for url data fetching,
    //   },
    // },
    table: {
      class: Table,
      inlineToolbar: true,
      config: {
        rows: 2,
        cols: 3,
        maxRows: 5,
        maxCols: 5,
      },
    },
    paragraph: {
      class: Paragraph,
      inlineToolbar: false,
    },
    header: {
      class: Header,
      config: {
        placeholder: "Enter a header",
        levels: [2, 3, 4],
        defaultLevel: 3,
      },
      shortcut: "CMD+SHIFT+H",
    },
    quote: {
      class: Quote,
      inlineToolbar: true,
      shortcut: "CMD+SHIFT+O",
      config: {
        quotePlaceholder: "Enter a quote",
        captionPlaceholder: "Quote's author",
      },
    },
    list: {
      class: EditorjsList,
      inlineToolbar: true,
      config: {
        defaultStyle: "unordered",
      },
    },
  },
});
