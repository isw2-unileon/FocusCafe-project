import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from "react-router-dom";
import { BookOpen, Clock, Upload, CheckCircle2, Coffee, Brain } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';

type SessionState = 'SETUP' | 'STUDYING' | 'QUIZ' | 'RESULTS';

interface QuizQuestion {
    question: string;
    options: string[];
    correctAnswer: number;
}

const StudySession = () => {
    const { userStats, isAuthenticated } = useAuth();
    const navigate = useNavigate();

    // Form states:
    const [state, setState] = useState<SessionState>('SETUP');
    const [files, setFiles] = useState<FileList | null>(null);
    const [studyMinutes, setStudyMinutes] = useState(25);
    const [timeLeft, setTimeLeft] = useState(0);

    // Quiz states:
    const [quiz, setQuiz] = useState<QuizQuestion[]>([]);
    const [userAnswers, setUserAnswers] = useState<number[]>([]);
    const [isGenerating, setIsGenerating] = useState(false);
    const [currentSessionId, setCurrentSessionId] = useState<number | null>(null);

    useEffect(() => {
        if (!isAuthenticated) {
            navigate('/');
        }
    }, [isAuthenticated, navigate]);

    const formatTime = (seconds: number) => {
        const m = Math.floor(seconds / 60);
        const s = seconds % 60;
        return `${m}:${s < 10 ? '0' : ''}${s}`;
    };

    const handleStartStudy = async () => {
        if (!files || files.length === 0) {
            alert("Please upload at least one file to start studying.");
            return;
        }

        const selectedFile = files[0];
        if (!selectedFile) return;
        const formData = new FormData();
        formData.append('pdf', selectedFile);
        formData.append('subject_name', 'General Study');

        try {
            const response = await fetch('http://localhost:8081/api/study/start', {
                method: 'POST',
                body: formData,
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (!response.ok) throw new Error("Failed to start session");

            const data = await response.json();
            setCurrentSessionId(data.session_id);
            setTimeLeft(studyMinutes * 60);
            setState('STUDYING');
        } catch (error) {
            console.error("Error starting session:", error);
            alert("Server connection failed. Check if the backend is running.");
        }
    };

    const handleStartQuiz = useCallback(async () => {
        if (!currentSessionId) return;

        setIsGenerating(true);
        setState('QUIZ');

        try {
            const response = await fetch(`http://localhost:8081/api/study/generate-quiz/${currentSessionId}`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (!response.ok) throw new Error("AI failed to generate quiz");

            const data = await response.json();
            
            // Parsing logic to handle both array and string formats from the backend
            let parsedQuiz: QuizQuestion[] = [];
            
            if (Array.isArray(data.quiz)) {
                parsedQuiz = data.quiz;
            } else if (typeof data.quiz === 'string') {
                try {
                    // 1. Clean the string from potential markdown code block formatting
                    const cleanString = data.quiz.replace(/```json|```/g, "").trim();
                    
                    // 2. Attempt to parse the cleaned string as JSON
                    const rawData = JSON.parse(cleanString);
                    
                    // 3. Handle different possible structures (array directly, or nested under 'quiz' or 'questions')
                    if (Array.isArray(rawData)) {
                        parsedQuiz = rawData;
                    } else if (rawData.quiz && Array.isArray(rawData.quiz)) {
                        parsedQuiz = rawData.quiz;
                    } else if (rawData.questions && Array.isArray(rawData.questions)) {
                        parsedQuiz = rawData.questions;
                    }
                } catch (e) {
                    console.error("Failed to parse quiz string:", e);
                    // Fallback: Try to extract JSON array from the string using regex
                    const jsonMatch = data.quiz.match(/\[[\s\S]*\]/);
                    if (jsonMatch) parsedQuiz = JSON.parse(jsonMatch[0]);
                }
            }

            // Validar que realmente sea un array antes de setear
            setQuiz(Array.isArray(parsedQuiz) ? parsedQuiz : []);
            
        } catch (error) {
            console.error("Quiz generation error:", error);
            setQuiz([
                { 
                    question: "Hubo un problema al generar el cuestionario. ¿Deseas intentarlo de nuevo?", 
                    options: ["Reintentar"], 
                    correctAnswer: 0 
                }
            ]);
        } finally {
            setIsGenerating(false);
        }
    }, [currentSessionId]);

    const handleFinishQuiz = async () => {
        const correctAnswers = userAnswers.filter((ans, i) => quiz[i] && ans === quiz[i].correctAnswer).length;
        const pointsEarned = correctAnswers * 2;

        try {
            await fetch('http://localhost:8081/api/user/progress', {
                method: 'POST',
                headers: { 
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                },
                body: JSON.stringify({ 
                    points: pointsEarned,
                    session_id: currentSessionId
                })
            });
        } catch (e) {
            console.error("Error saving progress:", e);
        }

        setState('RESULTS');
    };

    useEffect(() => {
        let timer: NodeJS.Timeout;
        if (state === 'STUDYING' && timeLeft > 0) {
            timer = setInterval(() => setTimeLeft(prev => prev - 1), 1000);
        } else if (state === 'STUDYING' && timeLeft === 0) {
            handleStartQuiz();
        }
        return () => clearInterval(timer);
    }, [state, timeLeft, handleStartQuiz]);

    if (!userStats) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-stone-100">
                <div className="text-center">
                    <Coffee className="animate-bounce mx-auto text-orange-600 mb-4" size={40} />
                    <p className="font-bold text-stone-600">Loading your profile...</p>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-stone-100 p-6">
            <div className="max-w-3xl mx-auto">
                <div className="flex items-center gap-4 mb-8">
                    <button onClick={() => navigate('/home')} className="text-stone-500 hover:text-stone-800">← Back</button>
                    <h1 className="text-3xl font-black text-stone-800 flex items-center gap-2">
                        <Coffee className="text-orange-600" /> Study & Brew
                    </h1>
                </div>

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

                {state === 'QUIZ' && (
                    <div className="bg-white rounded-3xl p-8 shadow-sm">
                        {isGenerating ? (
                            <div className="text-center py-12">
                                <Brain className="mx-auto text-orange-50 animate-bounce mb-4" size={48} />
                                <h2 className="text-2xl font-black">AI is crafting your test...</h2>
                                <p className="text-stone-500">Analyzing your PDFs to check your knowledge.</p>
                            </div>
                        ) : (
                            <div className="space-y-8">
                                <h2 className="text-2xl font-black flex items-center gap-2"><BookOpen /> Evaluation Time</h2>
                                {quiz && quiz.length > 0 ? quiz.map((q, idx) => (
                                    <div key={idx} className="p-6 bg-stone-50 rounded-2xl border border-stone-100">
                                        <p className="font-bold text-lg mb-4">{idx + 1}. {q.question}</p>
                                        <div className="grid grid-cols-1 gap-3">
                                            {q.options.map((opt, i) => (
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
                                )) : <p className="text-center py-4">No questions available.</p>}
                                <button onClick={handleFinishQuiz} className="w-full bg-stone-900 text-white py-4 rounded-xl font-bold">SUBMIT ANSWERS</button>
                            </div>
                        )}
                    </div>
                )}

                {state === 'RESULTS' && (
                    <div className="bg-white rounded-[3rem] p-12 text-center shadow-2xl">
                        <div className="w-24 h-24 bg-green-100 text-green-600 rounded-full flex items-center justify-center mx-auto mb-6">
                            <CheckCircle2 size={48} />
                        </div>
                        <h2 className="text-4xl font-black mb-2">Session Complete!</h2>
                        <p className="text-stone-500 mb-8 font-medium">You scored {userAnswers.filter((ans, i) => quiz[i] && ans === quiz[i].correctAnswer).length} / {quiz.length}</p>
                        <button onClick={() => navigate('/home')} className="bg-stone-900 text-white px-8 py-4 rounded-xl font-bold hover:bg-stone-800 transition-all">RETURN TO CAFETERIA</button>
                    </div>
                )}
            </div>
        </div>
    );
};

export default StudySession;