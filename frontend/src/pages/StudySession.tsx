import React, { useEffect, useState } from 'react';
import {useNavigate} from "react-router-dom";
import {BookOpen, Clock, Upload, CheckCircle2, Coffee, Brain} from 'lucide-react';
import { useGame } from "@/context/useGame";

type SessionState = 'SETUP' |'STUDYING'|'QUIZ'|'RESULTS';

interface QuizQuestion {
    question: string;
    options: string[];
    correct: number;
}

const StudySession = () => {
    const {user, addXP} = useGame();
    const navigate = useNavigate();

    // Form states:
    const [state, setState] = useState<SessionState>('SETUP');
    const [files, setFiles] = useState<FileList | null>(null);
    const [studyMinutes, setStudyMinutes] = useState(25); // Default to 25 minutes
    const [timeLeft, setTimeLeft] = useState(0);

    // Quiz states:
    const [quiz, setQuiz] = useState<QuizQuestion[]>([]);
    const [userAnswers, setUserAnswers] = useState<number[]>([]);
    const [isGenerating, setIsGenerating] = useState(false);

    // Timer logic:
    useEffect(() => {
        let timer: NodeJS.Timeout;
        if(state === 'STUDYING' && timeLeft > 0) {
            timer = setInterval(() => setTimeLeft(prev => prev - 1), 1000); // Decrease time every second
        } else if(state === 'STUDYING' && timeLeft === 0) {
            handleStartQuiz();
        }
        return () => clearInterval(timer);
    }, [state, timeLeft]);    
    
    const formatTime = (seconds: number) => {
        const m = Math.floor(seconds / 60);
        const s = seconds % 60;
        return `${m}:${s < 10 ? '0' : ''}${s}`;
    };

    const handleStartStudy = () => {
        if(!files || files.length === 0) {
            alert("Please upload at least one file to start studying.");
            return;
        }
        setTimeLeft(studyMinutes * 60); // Convert minutes to seconds
        setState('STUDYING'); // Start the study session
    };

    const handleStartQuiz = async () => {
        setIsGenerating(true);
        setState('QUIZ');

        //Mocking the Backend Call to Go API
        // This would be: fetch('/api/generate-quiz', { method: 'POST', body: formData })
        setTimeout(() => {
            // Mock quiz data
            setQuiz([
                { question: "What is the primary concept of the uploaded text?", options: ["Option A", "Option B", "Option C", "Option D"], correct: 0 },
                { question: "Based on page 2, why is X important?", options: ["Reason 1", "Reason 2", "Reason 3", "Reason 4"], correct: 2 },
            ]);
            setIsGenerating(false);
        }, 2000);
    };

    const handleFinishQuiz = () => {
        const correctAnswers = userAnswers.filter((ans, i) => quiz[i] && ans === quiz[i].correct).length;
        const score = correctAnswers / quiz.length;

        if (score >= 0.75) { // If user scores 75% or higher, they earn XP. It could be adjusted based on difficulty or other factors.
            addXP(100); // Award 100 XP for passing the quiz. 
        }
        setState('RESULTS');
    };

    return ( // Overall container with padding and background color
        <div className="min-h-screen bg-stone-100 p-6">
            <div className="max-w-3xl mx-auto">
                {/* Header */}
                <div className="flex items-center gap-4 mb-8">
                    <button onClick={() => navigate('/home')} className="text-stone-500 hover:text-stone-800">← Back</button>
                    <h1 className="text-3xl font-black text-stone-800 flex items-center gap-2">
                        <Coffee className="text-orange-600" /> Study & Brew
                    </h1>
                </div>

                {/* State: SETUP */}
                {state === 'SETUP' && (
                    <div className="bg-white rounded-3xl p-8 shadow-sm border border-stone-200">
                        <h2 className="text-xl font-bold mb-6">Prepare your Session</h2>
                        
                        <div className="space-y-6">
                            <div>
                                <label className="block text-sm font-black text-stone-500 uppercase mb-2">1. Upload Material (PDF)</label>
                                <div className="border-2 border-dashed border-stone-200 rounded-2xl p-8 text-center hover:border-orange-400 transition-colors cursor-pointer relative">
                                    <input type="file" multiple accept=".pdf" onChange={(e) => setFiles(e.target.files)} className="absolute inset-0 opacity-0 cursor-pointer" />
                                    <Upload className="mx-auto text-stone-400 mb-2" />
                                    <p className="text-stone-600 font-medium">{files ? `${files.length} files selected` : "Drag and drop your study PDFs here"}</p>
                                </div>
                            </div>

                            <div>
                                <label className="block text-sm font-black text-stone-500 uppercase mb-2">2. Focus Time (Minutes)</label>
                                <input type="number" value={studyMinutes} onChange={(e) => setStudyMinutes(Number(e.target.value))} className="w-full bg-stone-50 border border-stone-200 rounded-xl p-4 text-xl font-bold focus:ring-2 focus:ring-orange-500 outline-none" />
                            </div>

                            <button onClick={handleStartStudy} className="w-full bg-stone-900 text-white py-6 rounded-2xl font-black text-xl hover:bg-orange-600 transition-all shadow-lg active:scale-95">
                                START BREWING
                            </button>
                        </div>
                    </div>
                )}

                {/* State: STUDYING */}
                {state === 'STUDYING' && (
                    <div className="text-center py-20 bg-white rounded-[3rem] shadow-xl border-8 border-orange-50">
                        <div className="relative inline-block">
                            <Clock size={120} className="text-stone-100 absolute -inset-4 animate-pulse" />
                            <h2 className="text-7xl font-black text-stone-800 relative">{formatTime(timeLeft)}</h2>
                        </div>
                        <p className="text-stone-500 mt-8 font-medium italic">"Focus on the material. The quiz starts when the timer ends."</p>
                        <button onClick={() => setTimeLeft(0)} className="mt-12 text-stone-400 hover:text-orange-600 text-sm font-bold uppercase tracking-widest">Skip to Quiz (Debug)</button>
                    </div>
                )}

                {/* State: QUIZ */}
                {state === 'QUIZ' && (
                    <div className="bg-white rounded-3xl p-8 shadow-sm">
                        {isGenerating ? (
                            <div className="text-center py-12">
                                <Brain className="mx-auto text-orange-500 animate-bounce mb-4" size={48} />
                                <h2 className="text-2xl font-black">IA is crafting your test...</h2>
                                <p className="text-stone-500">Analysing your PDFs to check your knowledge.</p>
                            </div>
                        ) : (
                            <div className="space-y-8">
                                <h2 className="text-2xl font-black flex items-center gap-2"><BookOpen /> Evaluation Time</h2>
                                {quiz.map((q, idx) => (
                                    <div key={idx} className="p-6 bg-stone-50 rounded-2xl border border-stone-100">
                                        <p className="font-bold text-lg mb-4">{idx + 1}. {q.question}</p>
                                        <div className="grid grid-cols-1 gap-3">
                                            {q.options.map((opt: string, i: number) => (
                                                <button key={i} onClick={() => {
                                                    const newAns = [...userAnswers];
                                                    newAns[idx] = i;
                                                    setUserAnswers(newAns);
                                                }} className={`p-4 rounded-xl text-left font-medium transition-all ${userAnswers[idx] === i ? 'bg-orange-600 text-white shadow-md' : 'bg-white hover:bg-stone-100'}`}>
                                                    {opt}
                                                </button>
                                            ))}
                                        </div>
                                    </div>
                                ))}
                                <button onClick={handleFinishQuiz} className="w-full bg-stone-900 text-white py-4 rounded-xl font-bold">SUBMIT ANSWERS</button>
                            </div>
                        )}
                    </div>
                )}

                {/* State: RESULTS */}
                {state === 'RESULTS' && (
                    <div className="bg-white rounded-[3rem] p-12 text-center shadow-2xl">
                        <div className="w-24 h-24 bg-green-100 text-green-600 rounded-full flex items-center justify-center mx-auto mb-6">
                            <CheckCircle2 size={48} />
                        </div>
                        <h2 className="text-4xl font-black mb-2">Session Complete!</h2>
                        <p className="text-stone-500 mb-8 font-medium">You scored {userAnswers.filter((ans, i) => quiz[i] && ans === quiz[i].correct).length} / {quiz.length}</p>
                        
                        <div className="bg-orange-50 border border-orange-100 p-6 rounded-2xl mb-8">
                            <p className="text-orange-800 font-bold text-lg">+100 XP Earned!</p>
                            <p className="text-orange-600 text-sm">Keep it up to reach Level {user!.level + 1}</p>
                        </div>

                        <button onClick={() => navigate('/home')} className="bg-stone-900 text-white px-8 py-4 rounded-xl font-bold hover:bg-stone-800 transition-all">RETURN TO CAFETERIA</button>
                    </div>
                )}
            </div>
        </div>
    );
};

export default StudySession;