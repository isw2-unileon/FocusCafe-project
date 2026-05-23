import { Brain, CheckCircle2 } from 'lucide-react';

export interface QuizQuestion {
    question: string;
    options: string[];
    correctAnswer: number;
    explanation?: string;
}

interface StudyQuizProps {
    quiz: QuizQuestion[];
    isGenerating: boolean;
    userAnswers: number[];
    setUserAnswers: (answers: number[]) => void;
    onFinish: () => void;
    state: 'QUIZ' | 'RESULTS';
    onReturn: () => void;
}

export const StudyQuiz = ({ 
    quiz, isGenerating, userAnswers, setUserAnswers, onFinish, state, onReturn 
}: StudyQuizProps) => {
    
    const score = userAnswers.filter((ans, i) => quiz[i] && ans === quiz[i].correctAnswer).length;

    if (isGenerating) {
        return (
            <div className="text-center py-12 bg-white rounded-3xl shadow-sm">
                <Brain className="mx-auto text-orange-600 animate-bounce mb-4" size={48} />
                <h2 className="text-2xl font-black">AI is crafting your test...</h2>
            </div>
        );
    }

    if (state === 'RESULTS') {
        return (
            <div className="bg-white rounded-[3rem] p-12 text-center shadow-2xl border-8 border-green-50">
                <div className="w-24 h-24 bg-green-100 text-green-600 rounded-full flex items-center justify-center mx-auto mb-6">
                    <CheckCircle2 size={48} />
                </div>
                <h2 className="text-4xl font-black mb-2">Session Complete!</h2>
                <p className="text-stone-500 mb-8 text-xl">
                    Has acertado <span className="font-bold text-stone-800">{score}</span> de <span className="font-bold text-stone-800">{quiz.length}</span> preguntas.
                </p>
                <button onClick={onReturn} className="bg-stone-900 text-white px-8 py-4 rounded-xl font-bold hover:bg-orange-600 transition-colors">
                    RETURN TO CAFETERIA
                </button>
            </div>
        );
    }

    return (
        <div className="space-y-8 bg-white rounded-3xl p-8 shadow-sm">
            {quiz.map((q, idx) => (
                <div key={idx} className="p-6 bg-stone-50 rounded-2xl border border-stone-100">
                    <p className="font-bold text-lg mb-4">{idx + 1}. {q.question}</p>
                    <div className="grid grid-cols-1 gap-3">
                        {q.options.map((opt: string, i: number) => (
                            <button 
                                key={i} 
                                onClick={() => {
                                    const newAns = [...userAnswers];
                                    newAns[idx] = i;
                                    setUserAnswers(newAns);
                                }} 
                                className={`p-4 rounded-xl text-left transition-all ${
                                    userAnswers[idx] === i 
                                    ? 'bg-orange-600 text-white shadow-md' 
                                    : 'bg-white border border-stone-200 hover:border-orange-300'
                                }`}
                            >
                                {opt}
                            </button>
                        ))}
                    </div>
                </div>
            ))}
            <button 
                onClick={onFinish} 
                disabled={userAnswers.length < quiz.length}
                className="w-full bg-stone-900 text-white py-4 rounded-xl font-bold disabled:opacity-50 hover:bg-orange-600"
            >
                SUBMIT ANSWERS
            </button>
        </div>
    );
};