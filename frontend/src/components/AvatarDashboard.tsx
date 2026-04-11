import { GraduationCap } from "lucide-react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";

export const AvatarDashboard= ()=>{
    return (
        <div className="flex items-center gap-3">
            <Avatar className="h-11 w-11 border-2 shadow-sm transition-transform hover:scale-105">
                <AvatarFallback className="bg-white-100">
                    <GraduationCap
                        size={25}
                        className="text-orange-700"
                        strokeWidth={2}
                    />
                </AvatarFallback>
            </Avatar>
        </div>
    )
}