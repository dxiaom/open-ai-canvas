import { useEffect, useState } from "react";
import { Bot, Camera, Cat, CircleUserRound, Clapperboard, Rabbit, Rocket, Sparkles, type LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils";
import { isDefaultUserAvatarID, type DefaultUserAvatarID } from "@/lib/default-user-avatars";
import { resolveUserAvatarURL } from "@/lib/user-avatar-url";
import { apiBaseURL } from "@/services/api/request";
import type { LocalUser } from "@/stores/use-user-store";

const defaultAvatarIcons: Record<DefaultUserAvatarID, LucideIcon> = {
    person: CircleUserRound,
    director: Clapperboard,
    camera: Camera,
    cat: Cat,
    rabbit: Rabbit,
    robot: Bot,
    rocket: Rocket,
    sparkles: Sparkles,
};

export function UserAvatar({ user, className, iconClassName }: { user: LocalUser; className?: string; iconClassName?: string }) {
    const [failed, setFailed] = useState(false);
    const avatarUrl = user.avatarMode === "default" ? "" : resolveUserAvatarURL(user.avatarUrl, apiBaseURL);
    const DefaultAvatarIcon = defaultAvatarIcons[isDefaultUserAvatarID(user.avatarPreset) ? user.avatarPreset : "person"];

    useEffect(() => setFailed(false), [avatarUrl]);

    // 结构保持 button > span > svg：占位图标自动复用顶栏图标按钮的 18px/1.8 描边与配色合同；头像统一圆形，由组件兜底 rounded-full。
    // 图片用绝对定位铺满整个圆形容器（含 padding 区），避免百分比尺寸在 grid 自动行高下无法撑满而露出矩形边缘。
    return (
        <span className={cn("relative grid shrink-0 place-items-center overflow-hidden rounded-full", className)}>
            {avatarUrl && !failed ? (
                <img src={avatarUrl} alt="" referrerPolicy="no-referrer" className="absolute inset-0 size-full object-cover" onError={() => setFailed(true)} />
            ) : (
                <DefaultAvatarIcon className={cn("size-full", iconClassName)} aria-hidden />
            )}
        </span>
    );
}
