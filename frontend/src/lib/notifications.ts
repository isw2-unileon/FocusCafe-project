import Swal from 'sweetalert2';

export const showLevelUpModal = (newLevel: number) => {
    Swal.fire({
        title: 'LEVEL UP!',
        html: `
            <div class="flex flex-col items-center gap-4 py-4">
                <div class="relative">
                    <div class="text-7xl mb-2 animate-bounce">☕</div>
                    <div class="absolute -top-2 -right-2 text-4xl animate-pulse">✨</div>
                </div>
                <div class="text-center">
                    <p class="text-stone-600 font-medium mb-1">Congratulations! You've reached</p>
                    <h3 class="text-4xl font-black text-orange-600 tracking-tight">LEVEL ${newLevel}</h3>
                </div>
                <p class="text-sm text-stone-400 italic mt-2">New treats are waiting for you in the menu!</p>
            </div>
        `,
        confirmButtonText: 'AWESOME!',
        confirmButtonColor: '#ea580c', // orange-600
        buttonsStyling: false,
        customClass: {
            confirmButton: 'px-8 py-3 rounded-xl font-black text-white bg-orange-600 hover:bg-orange-700 transition-all active:scale-95 shadow-lg uppercase tracking-widest text-sm',
            popup: 'rounded-[2rem] border-8 border-white shadow-2xl',
            title: 'text-2xl font-black text-stone-800 pt-8'
        },
        backdrop: `rgba(41, 37, 36, 0.4)` // stone-900 with opacity
    });
};

export const showXpToast = (xp: number) => {
    Swal.fire({
        html: `
            <div class="flex items-center gap-3 px-2">
                <div class="bg-blue-100 p-2 rounded-lg">
                    <span class="text-xl">🏆</span>
                </div>
                <div class="flex flex-col items-start">
                    <span class="text-blue-600 font-black text-lg">+${xp} XP</span>
                    <span class="text-stone-500 text-[10px] font-bold uppercase tracking-wider">Experience gained</span>
                </div>
            </div>
        `,
        toast: true,
        position: 'top-end',
        showConfirmButton: false,
        timer: 3000,
        timerProgressBar: true,
        background: '#fff',
        customClass: {
            popup: 'rounded-2xl border-2 border-blue-50 shadow-xl p-4'
        }
    });
};

export const showOrderServedToast = () => {
    Swal.fire({
        html: `
            <div class="flex items-center gap-3 px-2">
                <div class="bg-orange-100 p-2 rounded-lg">
                    <span class="text-xl">☕</span>
                </div>
                <div class="flex flex-col items-start">
                    <span class="text-orange-600 font-black text-lg">ORDER SERVED!</span>
                    <span class="text-stone-500 text-[10px] font-bold uppercase tracking-wider">Freshly brewed just for you</span>
                </div>
            </div>
        `,
        toast: true,
        position: 'top-end',
        showConfirmButton: false,
        timer: 3000,
        timerProgressBar: true,
        background: '#fff',
        customClass: {
            popup: 'rounded-2xl border-2 border-orange-50 shadow-xl p-4'
        }
    });
};
