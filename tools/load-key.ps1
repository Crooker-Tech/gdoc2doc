<#
.SYNOPSIS
    Load an API key from Windows Credential Manager with cross-project fallback.
.PARAMETER Service
    Service name (e.g., together, google)
.PARAMETER Target
    Target usage (e.g., API, OAUTH) - used in env var name
.DESCRIPTION
    1. Try loading from current project's credential path
    2. If not found, search all credentials for matching /<service>-<target>-key suffix
    3. If found elsewhere, copy to current project's path for future use
    4. Only tries first matching fallback (no wasted time on bad keys)
.EXAMPLE
    . .\load-key.ps1 -Service together -Target API
    . .\load-key.ps1 -Service google -Target OAUTH
#>
param(
    [Parameter(Mandatory=$true)][string]$Service,
    [Parameter(Mandatory=$true)][string]$Target
)

$ProjectName = Split-Path -Leaf (Get-Location)
$CredentialSuffix = "$($Service.ToLower())-$($Target.ToLower())-key"
$CredentialName = "$ProjectName/$CredentialSuffix"
$EnvVarName = "$($Service.ToUpper())_$($Target.ToUpper())_KEY"

Add-Type -TypeDefinition @"
using System;
using System.Collections.Generic;
using System.Runtime.InteropServices;

public class CredentialManager {
    [DllImport("advapi32.dll", SetLastError = true, CharSet = CharSet.Unicode)]
    public static extern bool CredRead(string target, int type, int flags, out IntPtr credential);

    [DllImport("advapi32.dll", SetLastError = true, CharSet = CharSet.Unicode)]
    public static extern bool CredEnumerate(string filter, int flags, out int count, out IntPtr credentials);

    [DllImport("advapi32.dll", SetLastError = true)]
    public static extern bool CredFree(IntPtr credential);

    [DllImport("advapi32.dll", SetLastError = true, CharSet = CharSet.Unicode)]
    public static extern bool CredWrite(ref CREDENTIAL credential, int flags);

    [StructLayout(LayoutKind.Sequential, CharSet = CharSet.Unicode)]
    public struct CREDENTIAL {
        public int Flags;
        public int Type;
        public string TargetName;
        public string Comment;
        public System.Runtime.InteropServices.ComTypes.FILETIME LastWritten;
        public int CredentialBlobSize;
        public IntPtr CredentialBlob;
        public int Persist;
        public int AttributeCount;
        public IntPtr Attributes;
        public string TargetAlias;
        public string UserName;
    }

    public static string GetCredential(string target) {
        IntPtr credPtr;
        if (!CredRead(target, 1, 0, out credPtr)) return null;
        try {
            CREDENTIAL cred = (CREDENTIAL)Marshal.PtrToStructure(credPtr, typeof(CREDENTIAL));
            if (cred.CredentialBlobSize > 0) {
                return Marshal.PtrToStringUni(cred.CredentialBlob, cred.CredentialBlobSize / 2);
            }
            return null;
        } finally {
            CredFree(credPtr);
        }
    }

    public static List<string> FindCredentialsBySuffix(string suffix) {
        var results = new List<string>();
        IntPtr credPtr;
        int count;
        if (CredEnumerate(null, 0, out count, out credPtr)) {
            IntPtr current = credPtr;
            for (int i = 0; i < count; i++) {
                IntPtr credentialPtr = Marshal.ReadIntPtr(current);
                CREDENTIAL cred = (CREDENTIAL)Marshal.PtrToStructure(credentialPtr, typeof(CREDENTIAL));
                if (cred.TargetName != null && cred.TargetName.EndsWith("/" + suffix)) {
                    results.Add(cred.TargetName);
                }
                current = IntPtr.Add(current, IntPtr.Size);
            }
            CredFree(credPtr);
        }
        return results;
    }

    public static void SaveCredential(string target, string secret) {
        byte[] byteArray = System.Text.Encoding.Unicode.GetBytes(secret);
        CREDENTIAL cred = new CREDENTIAL();
        cred.Type = 1;
        cred.TargetName = target;
        cred.CredentialBlobSize = byteArray.Length;
        cred.CredentialBlob = Marshal.AllocHGlobal(byteArray.Length);
        Marshal.Copy(byteArray, 0, cred.CredentialBlob, byteArray.Length);
        cred.Persist = 2;
        cred.UserName = "API_KEY";
        CredWrite(ref cred, 0);
        Marshal.FreeHGlobal(cred.CredentialBlob);
    }
}
"@ -ErrorAction SilentlyContinue

# Try loading from current project first
$key = [CredentialManager]::GetCredential($CredentialName)

if (-not $key) {
    Write-Host "Key not found at '$CredentialName', searching other projects..." -ForegroundColor Yellow

    # Find all credentials with matching suffix
    $matches = [CredentialManager]::FindCredentialsBySuffix($CredentialSuffix)

    if ($matches.Count -gt 0) {
        $sourceCred = $matches[0]  # Only try first match
        Write-Host "  Found: $sourceCred" -ForegroundColor Cyan

        $key = [CredentialManager]::GetCredential($sourceCred)

        if ($key) {
            # Copy to current project's path
            [CredentialManager]::SaveCredential($CredentialName, $key)
            Write-Host "  Copied to '$CredentialName'" -ForegroundColor Green
        }
    }
}

if ($key) {
    Set-Item -Path "env:$EnvVarName" -Value $key
    Write-Host "$EnvVarName loaded." -ForegroundColor Green
} else {
    Write-Error "No $Service $Target key found. Run: .\save-key.ps1 -Service $Service -Target $Target -Key <your_key>"
    exit 1
}
