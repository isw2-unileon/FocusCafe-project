import { useContext } from "react";
import { GameContext } from "./GameContext";

export const useGame = () => {
    const context = useContext(GameContext);
    console.log("useGame context value:", context); // Debug log to check context value
    if(!context){
        throw new Error("useGame must be used within a GameProvider");
    }
    return context;
}