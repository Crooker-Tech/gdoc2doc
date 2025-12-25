<#
.SYNOPSIS
    Store an API key in Windows Credential Manager.
.PARAMETER Service
    Service name (e.g., together, google)
.PARAMETER Target
    Target usage (e.g., API, OAUTH) - used in env var name
.PARAMETER Key
    The API key value to store
.EXAMPLE
    .\save-key.ps1 -Service together -Target API -Key "your_key_here"
    .\save-key.ps1 -Service google -Target OAUTH -Key "path/to/credentials.json"
#>
param(
    [Parameter(Mandatory=$true)][string]$Service,
    [Parameter(Mandatory=$true)][string]$Target,
    [Parameter(Mandatory=$true)][string]$Key
)

$ProjectName = Split-Path -Leaf (Get-Location)
$CredentialName = "$ProjectName/$($Service.ToLower())-$($Target.ToLower())-key"

Add-Type -TypeDefinition @"
using System;
using System.Runtime.InteropServices;
using System.Text;

public class CredentialManager {
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

    public static void SaveCredential(string target, string secret) {
        byte[] byteArray = Encoding.Unicode.GetBytes(secret);
        CREDENTIAL cred = new CREDENTIAL();
        cred.Type = 1;
        cred.TargetName = target;
        cred.CredentialBlobSize = byteArray.Length;
        cred.CredentialBlob = Marshal.AllocHGlobal(byteArray.Length);
        Marshal.Copy(byteArray, 0, cred.CredentialBlob, byteArray.Length);
        cred.Persist = 2;
        cred.UserName = "$($Target.ToUpper())_KEY";
        if (!CredWrite(ref cred, 0)) {
            throw new Exception("Failed to save credential");
        }
        Marshal.FreeHGlobal(cred.CredentialBlob);
    }
}
"@ -ErrorAction SilentlyContinue

[CredentialManager]::SaveCredential($CredentialName, $Key)
$EnvVarName = "$($Service.ToUpper())_$($Target.ToUpper())_KEY"
Write-Host "$EnvVarName stored as '$CredentialName'" -ForegroundColor Green
