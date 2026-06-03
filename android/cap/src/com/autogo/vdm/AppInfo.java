package com.autogo.vdm;

import android.content.Context;
import android.content.Intent;
import android.content.pm.ActivityInfo;
import android.content.pm.ApplicationInfo;
import android.content.pm.PackageInfo;
import android.content.pm.PackageManager;
import android.content.pm.ResolveInfo;
import android.os.Looper;

import java.lang.reflect.Method;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class AppInfo {
    public static void main(String[] args) {
        try {
            printInstalledApps();
            System.exit(0);
        } catch (Throwable t) {
            t.printStackTrace(System.err);
            System.exit(1);
        }
    }

    private static void printInstalledApps() throws Exception {
        Context context = getSystemContext();
        if (context == null) {
            throw new IllegalStateException("无法获取 Android system context");
        }

        PackageManager pm = context.getPackageManager();
        Map<String, String> launcherActivities = queryLauncherActivities(pm);
        List<ApplicationInfo> apps = pm.getInstalledApplications(PackageManager.MATCH_DISABLED_COMPONENTS);

        for (ApplicationInfo app : apps) {
            String packageName = safeString(app.packageName);
            String name = appLabel(pm, app);
            String activityName = safeString(launcherActivities.get(packageName));
            List<String> activities = queryPackageActivities(pm, packageName);
            System.out.println("{\"name\":\"" + jsonEscape(name) +
                    "\",\"packageName\":\"" + jsonEscape(packageName) +
                    "\",\"activityName\":\"" + jsonEscape(activityName) +
                    "\",\"activities\":" + jsonStringArray(activities) + "}");
        }
    }

    private static Map<String, String> queryLauncherActivities(PackageManager pm) {
        Map<String, String> result = new HashMap<String, String>();
        Intent intent = new Intent(Intent.ACTION_MAIN);
        intent.addCategory(Intent.CATEGORY_LAUNCHER);
        List<ResolveInfo> activities = pm.queryIntentActivities(intent, PackageManager.MATCH_DISABLED_COMPONENTS);
        for (ResolveInfo info : activities) {
            if (info == null || info.activityInfo == null) {
                continue;
            }
            String packageName = safeString(info.activityInfo.packageName);
            String activityName = safeString(info.activityInfo.name);
            if (packageName.length() == 0 || activityName.length() == 0 || result.containsKey(packageName)) {
                continue;
            }
            result.put(packageName, activityName);
        }
        return result;
    }

    private static List<String> queryPackageActivities(PackageManager pm, String packageName) {
        List<String> result = new ArrayList<String>();
        try {
            PackageInfo packageInfo = pm.getPackageInfo(packageName,
                    PackageManager.GET_ACTIVITIES | PackageManager.MATCH_DISABLED_COMPONENTS);
            if (packageInfo == null || packageInfo.activities == null) {
                return result;
            }
            for (ActivityInfo activity : packageInfo.activities) {
                if (activity == null) {
                    continue;
                }
                String name = safeString(activity.name);
                if (name.length() == 0 || contains(result, name)) {
                    continue;
                }
                result.add(name);
            }
        } catch (RuntimeException ignored) {
            // Keep the app row visible even if one package cannot expose activities.
        } catch (PackageManager.NameNotFoundException ignored) {
            // Package disappeared during query; keep the app row visible.
        }
        return result;
    }

    private static boolean contains(List<String> values, String target) {
        for (String value : values) {
            if (target.equals(value)) {
                return true;
            }
        }
        return false;
    }

    private static Context getSystemContext() throws Exception {
        Class<?> activityThreadClass = Class.forName("android.app.ActivityThread");
        Object activityThread = null;

        try {
            Method currentActivityThread = activityThreadClass.getDeclaredMethod("currentActivityThread");
            currentActivityThread.setAccessible(true);
            activityThread = currentActivityThread.invoke(null);
        } catch (NoSuchMethodException ignored) {
            // Older or customized Android builds may not expose this helper.
        }

        if (activityThread == null) {
            if (Looper.myLooper() == null) {
                Looper.prepare();
            }
            Method systemMain = activityThreadClass.getDeclaredMethod("systemMain");
            systemMain.setAccessible(true);
            activityThread = systemMain.invoke(null);
        }

        Method getSystemContext = activityThreadClass.getDeclaredMethod("getSystemContext");
        getSystemContext.setAccessible(true);
        return (Context) getSystemContext.invoke(activityThread);
    }

    private static String appLabel(PackageManager pm, ApplicationInfo app) {
        if (app == null) {
            return "";
        }
        try {
            CharSequence label = pm.getApplicationLabel(app);
            if (label != null) {
                String text = label.toString();
                if (text.length() > 0) {
                    return text;
                }
            }
        } catch (RuntimeException ignored) {
            // Fall back to package name so the row remains searchable and explicit.
        }
        return safeString(app.packageName);
    }

    private static String safeString(String value) {
        return value == null ? "" : value;
    }

    private static String jsonEscape(String value) {
        StringBuilder out = new StringBuilder();
        for (int i = 0; i < value.length(); i++) {
            char c = value.charAt(i);
            switch (c) {
                case '"':
                    out.append("\\\"");
                    break;
                case '\\':
                    out.append("\\\\");
                    break;
                case '\b':
                    out.append("\\b");
                    break;
                case '\f':
                    out.append("\\f");
                    break;
                case '\n':
                    out.append("\\n");
                    break;
                case '\r':
                    out.append("\\r");
                    break;
                case '\t':
                    out.append("\\t");
                    break;
                default:
                    if (c < 0x20) {
                        String hex = Integer.toHexString(c);
                        out.append("\\u");
                        for (int pad = hex.length(); pad < 4; pad++) {
                            out.append('0');
                        }
                        out.append(hex);
                    } else {
                        out.append(c);
                    }
                    break;
            }
        }
        return out.toString();
    }

    private static String jsonStringArray(List<String> values) {
        StringBuilder out = new StringBuilder();
        out.append('[');
        for (int i = 0; i < values.size(); i++) {
            if (i > 0) {
                out.append(',');
            }
            out.append('"').append(jsonEscape(values.get(i))).append('"');
        }
        out.append(']');
        return out.toString();
    }
}
