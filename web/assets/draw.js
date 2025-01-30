window.process = {
  env: {
    IS_PREACT: "false",
    NODE_ENV: "development",
  },
};

import React from "react";
import { Excalidraw } from "@excalidraw/excalidraw";

const App = () => {
  return (
    <div style={{ height: "500px" }}>
      <Excalidraw
        initialData={{
          elements: [
            {
              type: "rectangle",
              x: 100,
              y: 100,
              width: 200,
              height: 100,
              strokeColor: "#000000",
            },
          ],
          appState: {
            viewBackgroundColor: "#ffffff",
          },
        }}
      />
    </div>
  );
};

const excalidrawWrapper = document.getElementById("draw");
const root = ReactDOM.createRoot(excalidrawWrapper);
root.render(React.createElement(App));
