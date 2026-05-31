import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from "react-router-dom";
import { Upload, Coffee,} from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import { apiFetch } from '@/services/api_client'; 
import { StudyQuiz } from '@/components/StudyQuiz';

// --- TYPES ---
type SessionState = 'SETUP' | 'STUDYING' | 'QUIZ' | 'RESULTS';

interface QuizQuestion {
    question: string;
    options: string[];
    correctAnswer: number;
    explanation?: string;
}

interface RawAIQuestion {
    question_text?: string;
    question?: string;
    option_a: string;
    option_b: string;
    option_c: string;
    option_d: string;
    correct_answer: string;
    explanation?: string;
}

// --- SERVICES ---
const studyService = {
    startSession: (formData: FormData) => 
        apiFetch('/study/start', { method: 'POST', body: formData }),
    
    generateQuiz: (sessionId: number) => 
        apiFetch(`/study/generate-quiz/${sessionId}`, { method: 'POST' }),
    
    saveProgress: (sessionId: number, score: number) => 
        apiFetch('/user/progress', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ score, session_id: sessionId })
        })
};

// --- MAIN COMPONENT ---
const StudySession = () => {
    const { userStats, setUserStats, isAuthenticated } = useAuth();
    const navigate = useNavigate();

    const [state, setState] = useState<SessionState>('SETUP');
    const [files, setFiles] = useState<FileList | null>(null);
    const [studyMinutes, setStudyMinutes] = useState(25);
    const [timeLeft, setTimeLeft] = useState(0);
    const [quiz, setQuiz] = useState<QuizQuestion[]>([]);
    const [userAnswers, setUserAnswers] = useState<number[]>([]);
    const [isGenerating, setIsGenerating] = useState(false);
    const [currentSessionId, setCurrentSessionId] = useState<number | null>(null);
    const [earnedEnergy, setEarnedEnergy] = useState<number>(0);
    const [hasError, setHasError] = useState(false);

    useEffect(() => {
        if (!isAuthenticated) navigate('/');
    }, [isAuthenticated, navigate]);

    const formatTime = (seconds: number) => {
        const m = Math.floor(seconds / 60);
        const s = seconds % 60;
        return `${m}:${s < 10 ? '0' : ''}${s}`;
    };

    // --- LOGIC HANDLERS  ---
    const handleStartStudy = async () => {
        if (!files || files.length === 0) return alert("Please upload a file.");

    const selectedFile = files[0];
    if (!selectedFile) return alert("Please upload a file.");
    
    const formData = new FormData();
    
    formData.append('pdf', selectedFile);
    
    formData.append('subject_name', 'General Study');

    console.log("Sending file:", selectedFile.name, "Size:", selectedFile.size);

    try {
        const data = await studyService.startSession(formData) as { session_id: number };
        console.log("Session started successfully:", data);
        setCurrentSessionId(data.session_id);
        setTimeLeft(studyMinutes * 60);
        setState('STUDYING');
    } catch (error) {
        const errorMessage = error instanceof Error ? error.message : String(error);
        console.error("Error 400 detailed:", errorMessage);
        alert(`Error: ${errorMessage}.Check the console.`);
    }
    };

    const handleStartQuiz = useCallback(async () => {
        if (!currentSessionId) return;
        setIsGenerating(true);
        setHasError(false);
        setState('QUIZ');

        try {
            const data = await studyService.generateQuiz(currentSessionId) as { questions: RawAIQuestion[] };
            const parsedQuiz = (data.questions || []).map((q: RawAIQuestion) => {
                const mapping: Record<string, number> = { 'A': 0, 'B': 1, 'C': 2, 'D': 3 };
                return {
                    question: q.question_text || q.question || "Untitled Question",
                    options: [q.option_a, q.option_b, q.option_c, q.option_d].filter(Boolean),
                    correctAnswer: mapping[String(q.correct_answer).trim().toUpperCase()] ?? 0,
                    explanation: q.explanation
                };
            });
            setQuiz(parsedQuiz);
        } catch (error) {
            console.error("Quiz error:", error);
            setHasError(true);
        } finally {
            setIsGenerating(false);
        }
    }, [currentSessionId]);

    const handleFinishQuiz = async () => {
        const score = userAnswers.filter((ans, i) => quiz[i] && ans === quiz[i].correctAnswer).length;
        try {
            const data = await studyService.saveProgress(currentSessionId!, score) as { new_total?: number, energy_earned?: number };
            if (data.energy_earned !== undefined) {
                setEarnedEnergy(data.energy_earned);
            }
            if (userStats && data.new_total !== undefined) {
                setUserStats({ ...userStats, energy: data.new_total });
            }
        } catch (e) {
            console.error("Error saving score:", e);
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

    if (!userStats) return null;

    return (
        <div className="min-h-screen bg-stone-100 p-6">
            <div className="max-w-3xl mx-auto">
                <div className="flex items-center gap-4 mb-8">
                    <button onClick={() => navigate('/home')} className="text-stone-500 hover:text-stone-800">← Back</button>
                    <h1 className="text-3xl font-black text-stone-800 flex items-center gap-2">
                        <Coffee className="text-orange-600" /> Study & Brew
                    </h1>
                </div>

                {/* STUDY SETUP OR STUDYING */}
                {(state === 'SETUP' || state === 'STUDYING') && (
                    <div className="space-y-6">
                        {state === 'SETUP' ? (
                            <div className="bg-white rounded-3xl p-8 shadow-sm border border-stone-200">
                                <h2 className="text-xl font-bold mb-6">Prepare your Session</h2>
                                <div className="space-y-6">
                                    <div className="border-2 border-dashed border-stone-200 rounded-2xl p-8 text-center relative hover:bg-stone-50 transition-colors">
                                        <input type="file" accept=".pdf" onChange={(e) => setFiles(e.target.files)} className="absolute inset-0 opacity-0 cursor-pointer" />
                                        <Upload className="mx-auto text-stone-400 mb-2" />
                                        <p className="text-stone-600">{files && files.length > 0 ? files[0]?.name : "Upload your PDF"}</p>
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-sm font-bold text-stone-500 uppercase">Study Time (min)</label>
                                        <input type="number" value={studyMinutes} onChange={(e) => setStudyMinutes(Number(e.target.value))} className="w-full bg-stone-50 border rounded-xl p-4 font-bold focus:ring-2 ring-orange-500 outline-none" />
                                    </div>
                                    <button onClick={handleStartStudy} className="w-full bg-stone-900 text-white py-6 rounded-2xl font-black hover:bg-orange-600 transition-all shadow-lg shadow-stone-200">START BREWING</button>
                                </div>
                            </div>
                        ) : (
                            <div className="text-center py-20 bg-white rounded-[3rem] shadow-xl border-8 border-orange-50">
                                <h2 className="text-7xl font-black text-stone-800 tracking-tighter">{formatTime(timeLeft)}</h2>
                                <p className="text-stone-500 mt-8 italic text-lg">Brewing knowledge... stay focused!</p>
                                <button onClick={() => setTimeLeft(0)} className="mt-12 text-xs text-stone-300 hover:text-orange-500 uppercase font-bold tracking-widest transition-colors">Skip to Quiz</button>
                            </div>
                        )}
                    </div>
                )}

                {/* STUDY QUIZ (Extracted Component) */}
                {(state === 'QUIZ' || state === 'RESULTS') && (
                    <StudyQuiz 
                        quiz={quiz}
                        isGenerating={isGenerating}
                        userAnswers={userAnswers}
                        setUserAnswers={setUserAnswers}
                        onFinish={handleFinishQuiz}
                        state={state}
                        onReturn={() => navigate('/home')}
                        earnedEnergy={earnedEnergy}
                        hasError={hasError}
                        onRetry={handleStartQuiz}
                    />
                )}
            </div>
        </div>
    );
};

export default StudySession;